// Package proxyhandler is an http.Handler which allows the RequestURI to be
// rewritten per request received.
package proxyhandler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

const WorkerCapacity = 5

var workPool = make(chan bool, WorkerCapacity)

// ProxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type ProxyHandler struct {
	http.Handler
	DefaultHost *url.URL
	routes      map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)
}

// New creates allocates a zero-value ProxyHandler and returns its pointer. defaultProxiedServer
// will provide the default host for handling routes which are not defined in the proxy
func New(defaultProxiedServer string) (*ProxyHandler, error) {
	p := ProxyHandler{}
	u, err := url.Parse(defaultProxiedServer)
	if err != nil {
		return nil, errors.New("Invalid default proxy host:" + err.Error())
	}
	p.DefaultHost = u
	p.routes = (make(map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)))
	return &p, nil
}

func copyHeaders(destination, source http.Header) {
	for k := range source {
		destination.Del(k)
	}
	for k, vs := range source {
		for _, v := range vs {
			destination.Add(k, v)
		}
	}
}

func copyBody(dest io.Writer, source io.ReadCloser) {
	defer source.Close()
	if closer, ok := dest.(io.Closer); ok {
		defer closer.Close()
	}
	_, err := io.Copy(dest, source)
	if err != nil {
		log.Printf("transfer body: %v", err.Error())
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

	pipeReader, pipeWriter := io.Pipe()
	go copyBody(pipeWriter, r.Body)

	req, err := http.NewRequest(r.Method, proxyRequestURL.String(), pipeReader)
	if err != nil {
		log.Printf("request: %v", err.Error())
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
	defer resp.Body.Close()
	if err != nil {
		log.Printf("http request:", err.Error())
	}
	copyHeaders(w.Header(), resp.Header)
	copyBody(w, resp.Body)

	<-workPool
	complete <- true
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for regexp, h := range h.routes {
		if regexp.Match(([]byte)(r.URL.Host + r.URL.Path)) {
			h(w, r)
			return
		}
	}
	prepareHandler(h.DefaultHost)(w, r)
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
