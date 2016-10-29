package main

import (
	"github.com/placer14/proxy_handler"
	"log"
	"net/http"
)

func main() {
	p := proxy_handler.NewProxyHandler()
	proxyEndpoints := map[string]string{
		"/":    "http://reddit.com",
		"/foo": "http://cnn.com",
	}
	for endpoint, override := range proxyEndpoints {
		p.HandleVia(endpoint, override)
	}

	log.Fatalln(http.ListenAndServe(":8080", p))
}
