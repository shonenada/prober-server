package main

import (
	"log"

	"github.com/shonenada/prober-server/prober"
	"github.com/shonenada/prober-server/status"
)

func main() {
	p, err := prober.BuildProber()

	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	err = p.Valid()
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	log.Printf("Prober server running in `%s` type; Probe Duration: %s", p.Type, p.Duration)
	if p.Type == "HTTP" {
		log.Printf("HTTP URL: %s; HTTP Timeout: %d", p.HTTPSettings.URL, p.HTTPSettings.Timeout)
	} else if p.Type == "TCP" {
		log.Printf("TCP Host: %s; TCP Port: %d", p.TCPSettings.Host, p.TCPSettings.Port)
	} else if p.Type == "UDP" {
		log.Printf("UDP Host: %s; UDP Port: %d", p.UDPSettings.Host, p.UDPSettings.Port)
	}

	go p.RunForver()

	server := status.MakeHTTPServer()
	server.ListenAndServe()
}
