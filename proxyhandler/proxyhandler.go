// Package proxyhandler is an http.Handler which allows the RequestURI to be
// rewritten per request received.
package proxyhandler

import (
	"fmt"
	"github.com/koding/websocketproxy"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// ProxyHandler implements http.Handler and will override portions of the request URI
// prior to completing the request.
type ProxyHandler struct {
	defaultHostURL *url.URL
	routes         []*validRouteRule
}

// New creates a valid ProxyHandler and returns its pointer. It will
// return an error if the defaultRouteRule is invalid
func New(config *Configuration) (*ProxyHandler, error) {
	validConfig, err := config.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %s", err.Error())
	}
	handler := ProxyHandler{
		defaultHostURL: validConfig.DefaultRoute,
		routes:         validConfig.Routes,
	}
	handler.announceSetup()
	return &handler, nil
}

func (handler *ProxyHandler) announceSetup() {
	log.Println("New proxy created")
	log.Printf("Default proxy backend %s", handler.defaultHostURL.String())
	for _, route := range handler.routes {
		log.Printf("\tRoute %s -> %s", route.Path, route.Endpoint)
	}
}

func (handler *ProxyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, route := range handler.routes {
		if strings.HasPrefix(request.URL.Path, route.Path) {
			switch route.EndpointURL.Scheme {
			case "ws":
				handler.handleWebsocketRequest(route.EndpointURL, writer, request)
			case "http":
				handler.handleHTTPRequest(route.EndpointURL, writer, request)
			}
			return
		}
	}
	handler.handleHTTPRequest(handler.defaultHostURL, writer, request)
}

func buildDownstreamRequestURL(upstreamRequestURL, routeRuleURL *url.URL) *url.URL {
	return &url.URL{
		Scheme:     routeRuleURL.Scheme,
		Host:       routeRuleURL.Host,
		Path:       upstreamRequestURL.Path,
		RawPath:    upstreamRequestURL.RawPath,
		ForceQuery: upstreamRequestURL.ForceQuery,
		RawQuery:   upstreamRequestURL.RawQuery,
	}
}

func (handler *ProxyHandler) handleWebsocketRequest(routeEndpointURL *url.URL, upstreamWriter http.ResponseWriter, upstreamRequest *http.Request) {
	websocketRequestBackend := func(r *http.Request) *url.URL {
		return buildDownstreamRequestURL(r.URL, routeEndpointURL)
	}
	websocketProxy := websocketproxy.WebsocketProxy{
		Backend:  websocketRequestBackend,
		Upgrader: websocketproxy.DefaultUpgrader,
	}
	// disable CORS same-origin verification within proxy
	websocketProxy.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	websocketProxy.ServeHTTP(upstreamWriter, upstreamRequest)
}

func (handler *ProxyHandler) handleHTTPRequest(routeEndpointURL *url.URL, upstreamWriter http.ResponseWriter, upstreamRequest *http.Request) {
	downstreamRequest, err := buildProxyRequest(upstreamRequest, routeEndpointURL)
	if err != nil {
		handleUnexpectedError(err, upstreamWriter)
		return
	}

	log.Printf("proxy: request %s -> %s %s", upstreamRequest.URL.String(), downstreamRequest.Method, downstreamRequest.URL.String())
	downstreamResponse, err := http.DefaultClient.Do(downstreamRequest)
	if err != nil {
		handleUnexpectedError(err, upstreamWriter)
		return
	}

	defer downstreamResponse.Body.Close()
	copyHeaders(upstreamWriter.Header(), downstreamResponse.Header)
	upstreamWriter.WriteHeader(downstreamResponse.StatusCode)
	io.Copy(upstreamWriter, downstreamResponse.Body)
}

func buildProxyRequest(upstreamRequest *http.Request, routeOverrideURL *url.URL) (*http.Request, error) {
	proxiedRequestURL := buildDownstreamRequestURL(upstreamRequest.URL, routeOverrideURL)
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
	log.Printf("proxy: http request error: %s", err.Error())
	header := writer.Header()
	header.Add("X-Error", fmt.Sprintf("unexpected error encountered: %s", err.Error()))
	writer.WriteHeader(500)
	writer.Write([]byte("error: " + err.Error()))
}

func copyHeaders(destination, source http.Header) {
	for headerKey, headerValues := range source {
		for _, headerValue := range headerValues {
			destination.Add(headerKey, headerValue)
		}
	}
}
