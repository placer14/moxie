package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
)

var numberOfProxyWorkers = 5
var workPool = make(chan bool, numberOfProxyWorkers)

type ProxyHandler struct {
	http.Handler
	routes map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)
}

func NewProxyHandler() *ProxyHandler {
	p := &ProxyHandler{}
	(*p).routes = (make(map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)))
	return p
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for regexp, h := range h.routes {
		if regexp.Match(([]byte)(r.URL.Host + r.URL.Path)) {
			h(w, r)
		}
	}
}

func (p *ProxyHandler) handleVia(request, destination string) {
	h := func(w http.ResponseWriter, r *http.Request) {
		complete := make(chan bool)
		workPool <- true
		go proxyRequest(destination, w, complete)
		<-complete
	}
	r, err := regexp.Compile(request)
	if err != nil {
		log.Fatalln(err.Error())
	}
	p.routes[r] = h

}

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

func proxyRequest(destination string, w http.ResponseWriter, complete chan<- bool) {
	log.Println("Got request destination")
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
	log.Println("Workpool:", len(workPool), "Done ", destination)

	<-workPool
	complete <- true
}

func main() {
	p := NewProxyHandler()
	p.handleVia("/", "http://google.com")
	p.handleVia("/foo", "http://reddit.com")
	log.Fatalln(http.ListenAndServe(":8080", p))
}
