// Package proxyhandler is an http.Handler which allows the RequestURI to be
// rewritten per request received.
package proxyhandler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// proxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type proxyHandler struct {
	defaultHostURL *url.URL
	routes         []*route
}

// New creates allocates a zero-value proxyHandler and returns its pointer. It will
// return an error if defaultProxiedHost cannot be parsed by url.Parse()
func New(defaultProxiedHost string) (*proxyHandler, error) {
	handler := proxyHandler{}
	err := handler.setDefaultHostURL(defaultProxiedHost)
	if err != nil {
		return nil, err
	}
	handler.routes = make([]*route, 0, 0)
	return &handler, nil
}

// HandleEndpoint accepts a string `path` which is compared against incoming Requests.
// If the `path` matches the incoming Request.RequestURI, the Host value from `endpoint`
// is used in the resulting HTTP request instead. HandleEndpoint will return an error if
// the route is created with invalid arguments
//
// Each endpoint is considered in the same order they are registered to `proxyHandler`.
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
	route, err := newRoute(path, endpoint)
	if err != nil {
		return fmt.Errorf("proxy: error handling endpoint: %s", err.Error())
	}
	handler.routes = append(handler.routes, route)
	return nil
}

func (handler *proxyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, routeMap := range handler.routes {
		if strings.HasPrefix(request.URL.Path, routeMap.path) {
			handler.handleProxyRequest(routeMap.endpointURL, writer, request)
			return
		}
	}
	handler.handleProxyRequest(handler.defaultHostURL, writer, request)
}

func (handler *proxyHandler) handleProxyRequest(routeEndpointURL *url.URL, upstreamWriter http.ResponseWriter, upstreamRequest *http.Request) {
	downstreamRequest, err := buildProxyRequest(upstreamRequest, routeEndpointURL)
	if err != nil {
		handleUnexpectedError(err, upstreamWriter)
		return
	}

	log.Printf("proxy: request %s -> %s %s", upstreamRequest.URL.String(), downstreamRequest.Method, downstreamRequest.URL.String())

	downstreamResponse, err := http.DefaultClient.Do(downstreamRequest)
	if err != nil {
		handleUnexpectedError(fmt.Errorf(""), upstreamWriter)
		return
	}

	defer downstreamResponse.Body.Close()
	copyHeaders(upstreamWriter.Header(), downstreamResponse.Header)
	upstreamWriter.WriteHeader(downstreamResponse.StatusCode)
	io.Copy(upstreamWriter, downstreamResponse.Body)
}

func (handler *proxyHandler) setDefaultHostURL(host string) error {
	defaultHostURL, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("proxy: invalid default host: %s", err.Error())
	}
	if defaultHostURL.Host == "" {
		return fmt.Errorf("proxy: default host is missing hostname or ip")
	}
	handler.defaultHostURL = defaultHostURL
	return nil
}

func buildProxyRequest(upstreamRequest *http.Request, routeOverrideURL *url.URL) (*http.Request, error) {
	var schemeOverride string
	switch {
	case routeOverrideURL.Scheme != "":
		schemeOverride = routeOverrideURL.Scheme
	case upstreamRequest.URL.Scheme != "":
		schemeOverride = upstreamRequest.URL.Scheme
	default:
		schemeOverride = "http"
	}

	proxiedRequestURL := url.URL{
		Scheme:     schemeOverride,
		Host:       routeOverrideURL.Host,
		Path:       upstreamRequest.URL.Path,
		RawPath:    upstreamRequest.URL.RawPath,
		ForceQuery: upstreamRequest.URL.ForceQuery,
		RawQuery:   upstreamRequest.URL.RawQuery,
	}

	// Unsure how this might return an error as parts for proxiedRequestURL should be valid.
	proxyRequest, err := http.NewRequest(upstreamRequest.Method, proxiedRequestURL.String(), upstreamRequest.Body)
	if err != nil {
		return nil, err
	}
	copyHeaders(proxyRequest.Header, upstreamRequest.Header)
	return proxyRequest, nil
}

func handleUnexpectedError(err error, writer http.ResponseWriter) {
	// No test coverage here, beware regressions within
	log.Printf("http request: %v", err.Error())
	header := writer.Header()
	header.Add("X-Error", fmt.Sprintf("Unexpected proxied request failure: %s", err.Error()))
	writer.WriteHeader(500)
	writer.Write([]byte("Proxy error: " + err.Error()))
}

func copyHeaders(destination, source http.Header) {
	for headerKey, headerValues := range source {
		for _, headerValue := range headerValues {
			destination.Add(headerKey, headerValue)
		}
	}
}
