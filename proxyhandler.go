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

// proxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type proxyHandler struct {
	DefaultHost *url.URL
	routeMap    map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)
}

// New creates allocates a zero-value proxyHandler and returns its pointer. defaultProxiedServer
// will provide the default host for handling routes which are not defined in the proxy
func New(defaultProxiedServer string) (*proxyHandler, error) {
	handler := proxyHandler{}
	u, err := url.Parse(defaultProxiedServer)
	if err != nil {
		return nil, errors.New("proxy: invalid default host: " + err.Error())
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("proxy: invalid default host scheme")
	}
	handler.DefaultHost = u
	handler.routeMap = (make(map[*regexp.Regexp]func(http.ResponseWriter, *http.Request)))
	return &handler, nil
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

func proxyRequest(r *http.Request, proxyOverride *url.URL) (*http.Request, error) {
	proxyRequestURL := *(r.URL)
	if proxyOverride.Host != "" {
		proxyRequestURL.Host = proxyOverride.Host
	}
	if proxyRequestURL.Scheme == "" && r.URL.IsAbs() {
		proxyRequestURL.Scheme = "http"
	}

	// Errors on new Request can only fail on malformed URL and bad Method, both cases are
	// caught by the server which provided the original request from the client before
	// it was provided to the handler. The proxyOverride.Host is verified
	// on load to protect against potential malformed URLs as well. This should never error.
	proxiedRequest, err := http.NewRequest(r.Method, proxyRequestURL.String(), r.Body)
	if err != nil {
		return nil, err
	}
	copyHeaders(proxiedRequest.Header, r.Header)
	return proxiedRequest, nil
}

func handleUnexpectedHandlingError(err error, w http.ResponseWriter) {
	// No test coverage here, beware regressions within
	log.Printf("http request: %v", err.Error())
	w.WriteHeader(500)
	w.Header().Set("X-Error", "Unexpected proxied request failure:")
	w.Header().Add("X-Error", err.Error())
}

func (handler *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for regexp, proxiedRequestHandler := range handler.routeMap {
		if regexp.Match(([]byte)(r.URL.Host + r.URL.Path)) {
			proxiedRequestHandler(w, r)
			return
		}
	}
	prepareHandler(handler.DefaultHost)(w, r)
}

func prepareHandler(proxyOverride *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp *http.Response
		var err error

		proxiedRequest, err := proxyRequest(r, proxyOverride)
		if err != nil {
			handleUnexpectedHandlingError(err, w)
		}
		log.Printf("Got request %s\n\tAsking for %s\n", r.URL.String(), proxiedRequest.URL.String())
		c := &http.Client{}
		resp, err = c.Do(proxiedRequest)
		if err != nil {
			handleUnexpectedHandlingError(err, w)
		} else {
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			copyHeaders(w.Header(), resp.Header)
			copyBody(w, resp.Body)
		}
	}
}

// HandleEndpoint accepts a string `endpoint` which is compiled to a Regexp which is
// compared against incoming Requests. If the Regexp matches the incoming
// Request.RequestURI, the Host value from proxyOverride is used in the resulting
// HTTP request instead.
func (handler *proxyHandler) HandleEndpoint(endpoint string, proxyOverride *url.URL) {
	r, err := regexp.Compile(endpoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
	handler.routeMap[r] = prepareHandler(proxyOverride)
}
