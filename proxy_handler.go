package proxy_handler

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var numberOfProxyWorkers = 5
var workPool = make(chan bool, numberOfProxyWorkers)

type ProxyHandler struct {
	http.Handler
	routes map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)
}

func New() *ProxyHandler {
	p := &ProxyHandler{}
	(*p).routes = (make(map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)))
	return p
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for regexp, h := range h.routes {
		if regexp.Match(([]byte)(r.URL.Host + r.URL.Path)) {
			h(w, r)
			return
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

func proxyRequest(r *http.Request, proxyOverride *url.URL) *http.Request {
	proxyRequestUrl := *(r.URL)
	if proxyOverride.Host != "" {
		proxyRequestUrl.Host = proxyOverride.Host
	}
	if proxyRequestUrl.Scheme == "" && r.URL.IsAbs() {
		proxyRequestUrl.Scheme = "http"
	}
	proxyBody, err := ioutil.ReadAll(r.Body)
	req, err := http.NewRequest(r.Method, proxyRequestUrl.String(), strings.NewReader(string(proxyBody)))
	if err != nil {
		log.Println("Proxy error", err.Error())
	}
	return req
}

func handleRequest(w http.ResponseWriter, r *http.Request, proxyOverride *url.URL, complete chan<- bool) {
	var resp *http.Response
	var err error

	proxiedRequest := proxyRequest(r, proxyOverride)
	log.Printf("Got request %s\n\tAsking for %s\n", r.URL.String(), proxiedRequest.URL.String())
	c := &http.Client{}
	resp, err = c.Do(proxiedRequest)
	if err != nil {
		log.Println("Proxy error", err.Error())
	}
	defer resp.Body.Close()
	copyHeaders(w.Header(), resp.Header)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Fatalln("Proxy error", err.Error())
	}

	<-workPool
	complete <- true
}
