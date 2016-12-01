package main

import (
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

func main() {
	p, err := proxyhandler.New("http://0.0.0.0:8000")
	if err != nil {
		log.Fatal("Invalid proxied server URI")
	}

	log.Println("Starting proxy server on 0.0.0.0:8080...")
	log.Fatalln(http.ListenAndServe(":8080", p))
}
