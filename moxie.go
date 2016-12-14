package main

import (
	"flag"
	"fmt"
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

const defaultHostValue = "//http_three:8000"

var listenPort = flag.Int("port", 8080, "specify which port the proxy should listen on")
var defaultHost = flag.String("proxied-host", defaultHostValue, "default host to recieve proxied traffic")

var routes = []*proxyhandler.RouteRule{
	&proxyhandler.RouteRule{Path: "/foo", Endpoint: "//http_one:8001"},
	&proxyhandler.RouteRule{Path: "/bar", Endpoint: "//http_two:8002"},
	&proxyhandler.RouteRule{Path: "/ws", Endpoint: "//websocket_one", WebsocketEnabled: true},
}

func main() {
	flag.Parse()

	p, err := proxyhandler.New(*defaultHost)
	if err != nil {
		log.Fatalf("Error creating proxy: %s", err.Error())
	}

	log.Println("Building routing table...")
	for _, r := range routes {
		if err := p.HandleEndpoint(r); err != nil {
			log.Fatalln(err.Error())
		}
	}

	log.Printf("Listening on port %d...", *listenPort)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), p))
}
