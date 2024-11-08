// Package tsproxy implements a tailscale proxy in go.
// The listener is on port 443 and uses tailscale for setting up TLS.
package tsproxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"tailscale.com/tsnet"
)

// New creates a new tailscale node that proxies all traffic on hostname:443 to target.
func New(hostname, target string) (*proxyHandler, error) {
	remote, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("could not parse proxy target %s (try like https://example.com) ", target)
	}

	// Set up tailscale server, listen on addr.
	ts := &tsnet.Server{
		Hostname:  hostname,
		Ephemeral: true,
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
	return &proxyHandler{
		listener: ln,
		proxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(remote)
			},
			ModifyResponse: fixRedirects(remote.Host, fqdn),
		},
		closer: func() {
			ts.Close()
			ln.Close()
		},
	}, nil
}

type proxyHandler struct {
	listener net.Listener

	proxy  *httputil.ReverseProxy
	closer func()
}

// Start starts proxying requests from the tailnet to the proxy target.
func (ph *proxyHandler) Start() error {
	if ph.closer != nil {
		defer ph.closer()
	}
	return http.Serve(ph.listener, ph)
}

func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.proxy.ServeHTTP(w, r)
}

// fixRedirects fixes incorrect redirects from the proxied service in an http.ReverseProxy.
// If the service being proxied returns a 302 redirect containing fromURL, this function
// will replace it with toURL.
func fixRedirects(fromURL, toURL string) func(w *http.Response) error {
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

		if strings.Contains(location, fromURL) {
			w.Header.Set("Location",
				strings.Replace(location, fromURL, toURL, 1))
		}
		return nil
	}
}
