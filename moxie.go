package main

import (
	"flag"
	"fmt"
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

func main() {
	var listenPort = flag.Int("port", 8080, "specify which port the proxy should listen on")
	var defaultHost = flag.String("proxied-host", "http://http_three:8000", "default host to recieve proxied traffic")

	flag.Parse()

	var config = &proxyhandler.Configuration{
		DefaultRoute: *defaultHost,
		Routes: []*proxyhandler.RouteRule{
			&proxyhandler.RouteRule{Path: "/foo", Endpoint: "http://http_one:8001"},
			&proxyhandler.RouteRule{Path: "/bar", Endpoint: "http://http_two:8002"},
			&proxyhandler.RouteRule{Path: "/ws", Endpoint: "ws://websocket_one"},
		},
	}

	p, err := proxyhandler.New(config)
	if err != nil {
		log.Fatalf("Error creating proxy: %s", err.Error())
	}

	log.Printf("Listening on port %d...", *listenPort)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), p))
}
