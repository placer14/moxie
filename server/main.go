package main

import (
	"github.com/placer14/proxyhandler"
	"log"
	"net/http"
)

func main() {
	p, err := proxyhandler.New("http://msn.com")
	if err != nil {
		log.Fatal("Invalid proxied server URI")
	}

	err = p.HandleEndpoint("/foo", "http://cnn.com")
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Fatalln(http.ListenAndServe(":8080", p))
}
