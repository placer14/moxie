package main

import (
	"flag"
	"fmt"
	"github.com/placer14/moxie/proxyhandler"
	"log"
	"net/http"
)

var host = flag.String("host", "0.0.0.0", "default host to proxy traffic")
var port = flag.String("port", "8000", "default port to proxy traffic")

func main() {
	flag.Parse()
	log.Printf("Binding proxy server to %s:%s...", *host, *port)

	p, err := proxyhandler.New(fmt.Sprintf("http://%s:%s", *host, *port))
	if err != nil {
		log.Fatal("Invalid proxied server URI")
	}

	log.Println("Listening on port 8080...")
	log.Fatalln(http.ListenAndServe(":8080", p))
}
