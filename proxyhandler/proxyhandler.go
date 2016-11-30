// Package proxyhandler is an http.Handler which allows the RequestURI to be
// rewritten per request received.
package proxyhandler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// proxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type proxyHandler struct {
	DefaultHost    *url.URL
	DefaultHandler func(http.ResponseWriter, *http.Request)
	routeMap       map[string]func(http.ResponseWriter, *http.Request)
}

// New creates allocates a zero-value proxyHandler and returns its pointer. defaultProxiedServer
// will provide the default host for handling routes which are not defined in the proxy
func New(defaultProxiedHost string) (*proxyHandler, error) {
	handler := proxyHandler{}
	err := handler.setDefaultProxyHandler(defaultProxiedHost)
	if err != nil {
		return nil, err
	}
	handler.routeMap = (make(map[string]func(http.ResponseWriter, *http.Request)))
	return &handler, nil
}

// HandleEndpoint accepts a string `path` which is compared against incoming Requests.
// If the `path` matches the incoming Request.RequestURI, the Host value from `endpoint`
// is used in the resulting HTTP request instead.
func (handler *proxyHandler) HandleEndpoint(path, endpoint string) error {
	if len(path) == 0 {
		return errors.New("proxy: empty path")
	}
	overrideUrl, err := url.Parse(endpoint)
	if err != nil {
		return errors.New("proxy: invalid endpoint url: " + err.Error())
	}
	handler.routeMap[path] = prepareHandler(overrideUrl)
	return nil
}

func (handler *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for endpoint, proxiedRequestHandler := range handler.routeMap {
		if strings.HasPrefix(r.URL.Path, endpoint) {
			proxiedRequestHandler(w, r)
			return
		}
	}
	handler.DefaultHandler(w, r)
}

func (handler *proxyHandler) setDefaultProxyHandler(subject string) error {
	u, err := url.Parse(subject)
	if err != nil {
		return errors.New("proxy: invalid default host: " + err.Error())
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("proxy: invalid default host scheme")
	}
	handler.DefaultHost = u
	handler.DefaultHandler = prepareHandler(handler.DefaultHost)
	return nil
}

func prepareHandler(proxyOverride *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp *http.Response
		var err error

		proxyRequest, err := buildProxyRequest(r, proxyOverride)
		if err != nil {
			handleUnexpectedHandlingError(err, w)
			return
		}
		log.Printf("Got request %s\n\tAsking for %s", r.URL.String(), proxyRequest.URL.String())
		c := &http.Client{}
		resp, err = c.Do(proxyRequest)
		if err != nil {
			handleUnexpectedHandlingError(err, w)
			return
		} else {
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			copyHeaders(w.Header(), resp.Header)
			io.Copy(w, resp.Body)
		}
	}
}

func buildProxyRequest(r *http.Request, proxyOverride *url.URL) (*http.Request, error) {
	proxyRequestURL := *(r.URL)
	if proxyOverride.Host != "" {
		proxyRequestURL.Host = proxyOverride.Host
	}
	if proxyRequestURL.Scheme == "" {
		proxyRequestURL.Scheme = "http"
	}

	// Errors on new Request can only fail on malformed URL and bad Method, both cases are
	// caught by the server which provided the original request from the client before
	// it was provided to the handler. The proxyOverride.Host is verified
	// on load to protect against potential malformed URLs as well. This should never error.
	proxyRequest, err := http.NewRequest(r.Method, proxyRequestURL.String(), r.Body)
	if err != nil {
		return nil, err
	}
	copyHeaders(proxyRequest.Header, r.Header)
	return proxyRequest, nil
}

func handleUnexpectedHandlingError(err error, w http.ResponseWriter) {
	// No test coverage here, beware regressions within
	log.Printf("http request: %v", err.Error())
	w.WriteHeader(500)
	w.Header().Set("X-Error", "Unexpected proxied request failure:")
	w.Header().Add("X-Error", err.Error())
	w.Write([]byte("Proxy error: " + err.Error()))
}

func copyHeaders(destination, source http.Header) {
	for k, vs := range source {
		for _, v := range vs {
			destination.Add(k, v)
		}
	}
}
