package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"log"
	"flag"
)

var port = flag.Int("p", 8000, "port to listen on")

func main() {
	flag.Parse()
	log.Printf("Starting echo daemon on 0.0.0.0:%d...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), echoHandler{}))
}

type echoHandler struct{}
func (echo echoHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Printf("error printing request: %v", err.Error())
		return
	}
	log.Println(string(dump))
	response.Write(dump)
}
