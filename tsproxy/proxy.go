// Package tsproxy implements a tailscale proxy in go.
// The listener is on port 443 and uses tailscale for setting up TLS.
package tsproxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"tailscale.com/tsnet"
)

// New creates a new tailscale node that can proxy all traffic on hostname:443 to target.
func New(hostname string, target *url.URL, authKey string, controlURL string) (*ProxyHandler, error) {
	// Get auth key from the environment. If it's an OAuth client key, we'll have to create
	// an auth key for authenticating nodes.
	if strings.HasPrefix(authKey, "tskey-client-") {
		// TODO: handle this case.
		// Something like tailscale.NewClient(); client.CreateKey()
		log.Fatal("Sorry, OAuth keys are not yet supported. Generate an auth key at https://login.tailscale.com/admin/settings/keys")
	}

	// Set up tailscale server, listen on addr.
	cfgPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	ts := &tsnet.Server{
		Dir:        filepath.Join(cfgPath, "truenas-tailscale", hostname),
		Hostname:   hostname,
		AuthKey:    authKey,    // If blank, a login link will appear on the CLI.
		ControlURL: controlURL, // If blank, use tailscale - otherwise this should be a headscale control server.
	}

	// Bind to HTTPS port.
	ln, err := ts.ListenTLS("tcp", ":443")
	if err != nil {
		return nil, err
	}

	// Get the fully qualified tailnet name of this host.
	// This is required for fixing redirects from the proxied host.
	lc, err := ts.LocalClient()
	if err != nil {
		return nil, err
	}
	status, err := lc.StatusWithoutPeers(context.Background())
	if err != nil {
		return nil, err
	}
	if status.CurrentTailnet == nil {
		return nil, errors.New("getting fqdn: not connected")
	}
	fqdn := fmt.Sprintf("%s.%s", hostname, status.CurrentTailnet.MagicDNSSuffix)

	// Configure proxy
	h := ProxyHandler{
		listener: ln,
		Target:   *target,
		proxy:    &httputil.ReverseProxy{},
		closer: func() {
			ts.Close()
			ln.Close()
		},
	}
	h.proxy.Rewrite = h.rewrite
	h.proxy.ModifyResponse = h.fixRedirects(fqdn)
	return &h, nil
}

type ProxyHandler struct {
	listener net.Listener

	proxy  *httputil.ReverseProxy
	closer func()

	// Target can be updated while running.
	Target url.URL
}

// Start starts proxying requests from the tailnet to the proxy target.
func (ph *ProxyHandler) Start() error {
	if ph.closer != nil {
		defer ph.closer()
	}
	return http.Serve(ph.listener, ph)
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.proxy.ServeHTTP(w, r)
}

func (ph ProxyHandler) rewrite(r *httputil.ProxyRequest) {
	r.SetURL(&ph.Target)
}

// fixRedirects fixes incorrect redirects from the proxied service in an http.ReverseProxy.
// If the service being proxied returns a 302 redirect containing the host header it received,
// this function will replace it with rewriteTo.
func (ph ProxyHandler) fixRedirects(rewriteTo string) func(w *http.Response) error {
	return func(w *http.Response) error {
		if w.StatusCode != 302 {
			// Ignore anything other than temporary redirects.
			return nil
		}
		location := w.Header.Get("Location")
		if location == "" {
			// We only handle location header redirects, return if no header exists.
			return nil
		}

		if strings.Contains(location, ph.Target.Host) {
			w.Header.Set("Location",
				strings.Replace(location, ph.Target.Host, rewriteTo, 1))
		}
		return nil
	}
}
