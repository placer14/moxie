// Package proxyhandler is an http.Handler which allows the RequestURI to be
// rewritten per request received.
package proxyhandler

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

// ProxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type ProxyHandler struct {
	http.Handler
	routes map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)
}

// New creates allocates a zero-value ProxyHandler and returns its pointer
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

// HandleEndpoint accepts a string `endpoint` which is compiled to a Regexp which is
// compared against incoming Requests. If the Regexp matches the incoming
// Request.RequestURI, the Host value from proxyOverride is used in the resulting
// HTTP request instead.
func (h *ProxyHandler) HandleEndpoint(endpoint string, proxyOverride *url.URL) {
	r, err := regexp.Compile(endpoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
	h.routes[r] = prepareHandler(proxyOverride)
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
	proxyRequestURL := *(r.URL)
	if proxyOverride.Host != "" {
		proxyRequestURL.Host = proxyOverride.Host
	}
	if proxyRequestURL.Scheme == "" && r.URL.IsAbs() {
		proxyRequestURL.Scheme = "http"
	}
	proxyBody, err := ioutil.ReadAll(r.Body)
	req, err := http.NewRequest(r.Method, proxyRequestURL.String(), strings.NewReader(string(proxyBody)))
	if err != nil {
		log.Println("Proxy error", err.Error())
	}
	copyHeaders(req.Header, r.Header)
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
