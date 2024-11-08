package main

import (
	"flag"
	"log"

	"github.com/dwurf/truenas-tailscale/tsproxy"
)

var (
	hostname    = flag.String("hostname", "truenas-tailscale", "hostname to use in the tailnet")
	proxyTarget = flag.String("proxy-target", "http://127.0.0.1:80", "proxy to target requests to")
)

func main() {
	flag.Parse()

	log.Fatal(tsproxy.New(*hostname, *proxyTarget))
}
