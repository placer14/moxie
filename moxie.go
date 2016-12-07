package main

import (
	"flag"
	"fmt"
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

var listen_port = flag.String("port", "8080", "specify which port the proxy should listen on")
var default_host = flag.String("proxied-host", "0.0.0.0", "default host to recieve proxied traffic")
var default_port = flag.String("proxied-port", "8000", "default port to revieve proxied traffic")

func main() {
	flag.Parse()
	log.Printf("Pointing proxy server to default host %s:%s...", *default_host, *default_port)

	p, err := proxyhandler.New(fmt.Sprintf("http://%s:%s", *default_host, *default_port))
	if err != nil {
		log.Fatalf("Error creating proxy: %s", err.Error())
	}

	log.Printf("Listening on port %s...", *listen_port)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", *listen_port), p))
}
