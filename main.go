package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/dwurf/truenas-tailscale/truenas"
	"github.com/dwurf/truenas-tailscale/tsproxy"
	"github.com/hashicorp/go-retryablehttp"
)

// Check for apps every pollFrequency seconds.
const pollFrequency = 10

type config struct {
	truenasAPIKey     string
	truenasHostname   string
	tailscaleAPIKey   string
	tailscaleHostname string
}

func (cfg *config) parse() {
	defaultHostname, ok := os.LookupEnv("TRUENAS_HOSTNAME")
	if !ok {
		defaultHostname = "127.0.0.1"
	}

	flag.StringVar(&cfg.truenasAPIKey, "truenas-api-key", "", "TrueNAS API key (env: TRUENAS_API_KEY).")
	flag.StringVar(&cfg.truenasHostname, "truenas-hostname", defaultHostname, "TrueNAS hostname or IP (env: TRUENAS_HOSTNAME).")
	flag.StringVar(&cfg.tailscaleAPIKey, "tailscale-api-key", "", "Tailscale API Key (env: TS_AUTHKEY).")
	flag.StringVar(&cfg.tailscaleHostname, "tailscale-hostname", os.Getenv("TS_HOSTNAME"), "Hostname to use in the tailnet. Defaults to the hostname configured in TrueNAS (env: TS_HOSTNAME).")

	flag.Parse()

	// Configure this last to avoid showing the API key in the CLI help.
	if cfg.truenasAPIKey == "" {
		cfg.truenasAPIKey = os.Getenv("TRUENAS_API_KEY")
	}
}

func main() {
	var cfg config
	cfg.parse()
	truenasURL, err := url.Parse("http://" + cfg.truenasHostname)
	if err != nil {
		log.Fatalf("could not parse truenas IP %s (try TRUENAS_IP or TRUENAS_IP:HTTP_PORT", cfg.truenasHostname)
	}
	// Connect to TrueNAS REST API.
	wsEndpoint := fmt.Sprintf("%s/api/v2.0", truenasURL)
	client := truenas.NewClient(wsEndpoint, cfg.truenasAPIKey, httpClient())

	// If hostname is not configured, get it from the API.
	// This doubles as a connectivity check.
	hostname, err := client.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.tailscaleHostname == "" {
		cfg.tailscaleHostname = hostname
	}

	// Start a background process to monitor apps.
	proxies := make(proxySet)
	go func() {
		for {
			apps, err := client.Apps()
			if err != nil {
				log.Printf("Error fetching apps: %s", err)
			}

			for _, app := range apps {
				if app.State != truenas.StateRunning {
					// Ignore stopped apps.
					continue
				}

				if len(app.Portals) == 0 {
					// Cannot proxy if no portal is configured.
					continue
				}
				// All we want from the app is the port, the rest will be built from the truenas host details.
				proxyTarget, err := url.Parse(fmt.Sprintf("http://%s:%s", cfg.truenasHostname, app.Portals[0].Port()))
				if err != nil {
					log.Printf("Error parsing proxyTarget %s", proxyTarget)
					continue
				}
				proxies.ensure(app.Name, proxyTarget)
			}

			for name := range proxies {
				i := slices.IndexFunc(apps, func(app truenas.App) bool { return app.Name == name })
				if i < 0 {
					log.Printf("App %s removed, removing proxy.", name)
					delete(proxies, name)
				} else if apps[i].State != truenas.StateRunning {
					log.Printf("Removing proxy for app %s: state is %s.", name, apps[i].State)
					delete(proxies, name)
				}
			}

			time.Sleep(pollFrequency * time.Second)
		}
	}()

	// Reverse proxy TrueNAS on the TailScale network.
	proxy, err := tsproxy.New(cfg.tailscaleHostname, truenasURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(proxy.Start())
}

type proxySet map[string]*tsproxy.ProxyHandler

// ensure proxy is running.
func (p *proxySet) ensure(appName string, proxyTarget *url.URL) {
	if oldProxy, ok := (*p)[appName]; !ok {
		// Only supports the first portal for an app.
		proxy, err := tsproxy.New(appName, proxyTarget)
		if err != nil {
			log.Printf("Error creating proxy for %s -> %s: %s", appName, proxyTarget, err)
			return
		}
		log.Printf("Registering app %s, proxied to %s", appName, proxyTarget)
		(*p)[appName] = proxy
		go func() {
			log.Printf("%s: %s", appName, proxy.Start())
		}()
	} else if oldProxy.Target != *proxyTarget {
		log.Printf("Updating app %s, was proxied to %s, now %s", appName, &oldProxy.Target, proxyTarget)
		oldProxy.Target = *proxyTarget
	}
}

func httpClient() *http.Client {
	c := retryablehttp.NewClient()
	c.RetryMax = 10
	return c.StandardClient()
}
