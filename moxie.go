package main

import (
	"flag"
	"fmt"
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

const defaultHostValue = "defaulthost"
const defaultPortValue = 8000

var listenPort = flag.Int("port", 8080, "specify which port the proxy should listen on")
var defaultHost = flag.String("proxied-host", defaultHostValue, "default host to recieve proxied traffic")
var defaultPort = flag.Int("proxied-port", defaultPortValue, "default port to revieve proxied traffic")

type route struct{ path, endpoint string }

var routes = []route{
	route{path: "/foo", endpoint: "//backendone:8001"},
	route{path: "/bar", endpoint: "//backendtwo:8002"},
}

func main() {
	flag.Parse()

	log.Printf("Pointing proxy server to default host %s:%d...", *defaultHost, *defaultPort)
	p, err := proxyhandler.New(fmt.Sprintf("http://%s:%d", *defaultHost, *defaultPort))
	if err != nil {
		log.Fatalf("Error creating proxy: %s", err.Error())
	}

	log.Println("Building routing table...")
	for _, route := range routes {
		log.Printf("\troute %s -> %s", route.path, route.endpoint)
		if err := p.HandleEndpoint(route.path, route.endpoint); err != nil {
			log.Fatalln(err.Error())
		}
	}

	log.Printf("Listening on port %d...", *listenPort)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), p))
}
