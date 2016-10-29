package main

import (
	"github.com/placer14/proxy_handler"
	"log"
	"net/http"
	"net/url"
)

func main() {
	p := proxy_handler.NewProxyHandler()
	proxyEndpoints := map[string]string{
		"/":    "http://reddit.com",
		"/foo": "http://cnn.com",
	}
	for endpoint, override := range proxyEndpoints {
		if u, err := url.Parse(override); err == nil {
			p.HandleEndpoint(endpoint, u)
		} else {
			log.Printf("Unable to parse host from %s for endpoint %s", override, endpoint)
		}
	}

	log.Fatalln(http.ListenAndServe(":8080", p))
}
