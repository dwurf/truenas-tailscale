// Package tsproxy implements a tailscale proxy in go.
package tsproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

// New creates a new tailscale node that proxies all traffic on hostname to target.
func New(hostname, target string) error {
	remote, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("could not parse proxy target %s (try like https://example.com) ", target)
	}

	// Set up tailscale server, listen on addr
	s := new(tsnet.Server)
	s.Hostname = hostname
	s.Ephemeral = true
	defer s.Close()
	ln, err := s.Listen("tcp", ":80")
	if err != nil {
		return err
	}
	defer ln.Close()

	// Get a local client for the tailnet.
	lc, err := s.LocalClient()
	if err != nil {
		return err
	}

	h := proxyHandler{
		tsClient: lc,
		proxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(remote)
			},
		},
	}
	return http.Serve(ln, &h)
}

type proxyHandler struct {
	tsClient *tailscale.LocalClient
	proxy    *httputil.ReverseProxy
}

func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.proxy.ServeHTTP(w, r)
}
