package main

import (
	"github.com/placer14/proxyhandler"
	"log"
	"net/http"
	"net/url"
)

func main() {
	p, err := proxyhandler.New("http://msn.com")
	if err != nil {
		log.Fatal("Invalid proxied server URI")
	}
	proxyEndpoints := map[string]string{
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
