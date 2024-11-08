package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

var (
	addr        = flag.String("addr", ":80", "address to listen on")
	hostname    = flag.String("hostname", "truenas-tailscale", "hostname to use in the tailnet")
	proxyTarget = flag.String("proxy-target", "http://127.0.0.1:80", "proxy to target requests to")
)

func main() {
	flag.Parse()

	remote, err := url.Parse(*proxyTarget)
	if err != nil {
		log.Fatal("could not parse proxy target (try https://example.com) ", *proxyTarget)
	}

	// Set up tailscale server, listen on addr
	s := new(tsnet.Server)
	s.Hostname = *hostname
	s.Ephemeral = true
	defer s.Close()
	ln, err := s.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// Get a local client for the tailnet.
	lc, err := s.LocalClient()
	if err != nil {
		log.Fatal(err)
	}

	h := proxyHandler{
		tsClient: lc,
		proxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(remote)
			},
		},
	}
	log.Fatal(http.Serve(ln, &h))
}

func firstLabel(s string) string {
	s, _, _ = strings.Cut(s, ".")
	return s
}

type proxyHandler struct {
	tsClient *tailscale.LocalClient
	proxy    *httputil.ReverseProxy
}

func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	who, err := ph.tsClient.WhoIs(r.Context(), r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println("Got request", r.Method, r.URL, "from", who.UserProfile.LoginName, firstLabel(who.Node.ComputedName), r.RemoteAddr)
	ph.proxy.ServeHTTP(w, r)
}
