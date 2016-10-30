package proxy_handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
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
	prepareHandler(&url.URL{})(w, r)
}

func prepareHandler(proxyOverride *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		complete := make(chan bool)
		workPool <- true
		go handleRequest(w, r, proxyOverride, complete)
		<-complete
	}
}

func (p *ProxyHandler) HandleEndpoint(endpoint string, proxyOverride *url.URL) {
	r, err := regexp.Compile(endpoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
	p.routes[r] = prepareHandler(proxyOverride)
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

func handleRequest(w http.ResponseWriter, r *http.Request, proxyOverride *url.URL, complete chan<- bool) {
	var resp *http.Response
	var err error

	c := &http.Client{}
	proxyUrl := r.URL
	if proxyOverride.Host != "" {
		proxyUrl.Host = proxyOverride.Host
	}

	log.Printf("Got request %s\n\tAsking for %s\n", r.URL.String(), proxyUrl.String())
	switch r.Method {
	case http.MethodGet:
		resp, err = c.Get(proxyUrl.String())
	default:
		log.Println("Unknown method", r.Method)
	}

	if err != nil {
		log.Fatalln("Request:", err.Error())
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	bodyLen, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Fatalln("Body:", err.Error())
	}

	log.Printf("Completed request %s. Wrote %d bytes", r.URL.String(), bodyLen)
	<-workPool
	complete <- true
}
