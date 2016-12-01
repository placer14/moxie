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
	DefaultHost *url.URL
	endpointMaps []*endpointMap
}

// New creates allocates a zero-value proxyHandler and returns its pointer. defaultProxiedServer
// will provide the default host for handling routes which are not defined in the proxy
func New(defaultProxiedHost string) (*proxyHandler, error) {
	handler := proxyHandler{}
	err := handler.setDefaultProxyHandler(defaultProxiedHost)
	if err != nil {
		return nil, err
	}
	handler.endpointMaps = make([]*endpointMap, 0, 0)
	return &handler, nil
}

// HandleEndpoint accepts a string `path` which is compared against incoming Requests.
// If the `path` matches the incoming Request.RequestURI, the Host value from `endpoint`
// is used in the resulting HTTP request instead. Each endpoint is considered in the
// same order they are registered to `proxyHandler`.
//
// Example:
//
// If you were to register two endpoints like so:
//
// handler.HandleEndpoint("/", "www.baz.com")
// handler.HandleEndpoint("/foo", "www.test.com")
//
// A request for `/foo` against the server using this handler would have the request
// proxied to `www.baz.com` instead of `www.test.com` because it was registered first.
func (handler *proxyHandler) HandleEndpoint(path, endpoint string) error {
	route, err := newEndpointMap(path, endpoint)
	if err != nil {
		return errors.New("proxy: error handling endpoint:" + err.Error())
	}
	handler.endpointMaps = append(handler.endpointMaps, route)
	return nil
}

func (handler *proxyHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	for _, routeMap := range handler.endpointMaps {
		if strings.HasPrefix(request.URL.Path, routeMap.path) {
			handler.handleProxyRequest(routeMap.endpointURL, response, request)
			return
		}
	}
	handler.handleProxyRequest(handler.DefaultHost, response, request)
}

func (handler *proxyHandler) handleProxyRequest(endpointURL *url.URL, response http.ResponseWriter, request *http.Request) {
	proxyRequest, err := buildProxyRequest(request, endpointURL)
	if err != nil {
		handleUnexpectedHandlingError(err, response)
		return
	}
	log.Printf("Got request %s\n\tAsking for %s", request.URL.String(), proxyRequest.URL.String())
	c := &http.Client{}
	resp, err := c.Do(proxyRequest)
	if err != nil {
		handleUnexpectedHandlingError(err, response)
		return
	} else {
		defer resp.Body.Close()
		response.WriteHeader(resp.StatusCode)
		copyHeaders(response.Header(), resp.Header)
		io.Copy(response, resp.Body)
	}
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
	return nil
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
