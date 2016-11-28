package main

import (
	"./proxyhandler"
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

	log.Println("Starting proxy server on 0.0.0.0:8080...")
	log.Fatalln(http.ListenAndServe(":8080", p))
}
