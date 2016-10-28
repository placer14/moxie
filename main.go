package main

import (
	"io"
	"log"
	"net/http"
)

func copyHeaders(dst, src http.Header) {
	for k := range src {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func handleVia(request, destination string) {
	h := func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(destination)
		if err != nil {
			log.Fatalln("Proxy: ", err.Error())
		}
		defer resp.Body.Close()
		copyHeaders(w.Header(), resp.Header)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Fatalln("Proxy: ", err.Error())
		}
	}
	http.HandleFunc(request, h)
}

func main() {
	handleVia("/", "http://google.com")
	handleVia("/foo", "http://reddit.com")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
