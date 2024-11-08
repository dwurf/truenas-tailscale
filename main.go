package main

import (
	"flag"
	"log"

	"github.com/dwurf/truenas-tailscale/tsproxy"
)

var (
	hostname = flag.String("hostname", "truenas-tailscale", "Hostname to use in the tailnet.")
	truenas  = flag.String("truenas-url", "http://127.0.0.1:80", "Base URL of the TrueNAS UI")
)

func main() {
	flag.Parse()

	// Reverse proxy TrueNAS on the TailScale network.
	proxy, err := tsproxy.New(*hostname, *truenas)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(proxy.Start())
}
