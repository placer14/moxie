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
func New(defaultEndpoint string) (*ProxyHandler, error) {
	route := &RouteRule{
		Path:     "*", // this non-empty path will satisfy validation, but is not used for routing
		Endpoint: defaultEndpoint,
	}
	validRoute, err := route.validate()
	if err != nil {
		return nil, fmt.Errorf("proxy: invalid default host: %s", err.Error())
	}
	log.Printf("Creating proxy server pointed at default backend %s...", validRoute.EndpointURL.String())
	handler := ProxyHandler{}
	handler.defaultHostURL = validRoute.EndpointURL
	handler.routes = make([]*validRouteRule, 0, 0)
	return &handler, nil
}

// HandleEndpoint accepts a *RouteRule which is used to build a route table. The
// RouteRules in this table are compared against incoming Requests as follows:
//
// If the `path` is found at the beginning of incoming Request.URL.Path, the Host value
// from `endpoint` is used in the resulting HTTP request instead. HandleEndpoint will
// return an error if the RouteRule is invalid
//
// The RouteRules are considered in the same order they are registered to `ProxyHandler`:
//
// If you were to register two endpoints like so:
//
// handler.HandleEndpoint(&RouteRule{ Path: "/", Endpoint: "//baz.com" })
// handler.HandleEndpoint(&RouteRule{ Path: "/foo", Endpoint: "//test.com" })
//
// A request for `/foo` against the server using this handler would have the request
// proxied to `baz.com` instead of `test.com`. This is because `/foo` contains
// the first rule's path `/` at the beginning and was registered before the rule containing
// the path `/foo`. It is recommended that you register more specific rules before rules
// with less specificity.
func (handler *ProxyHandler) HandleEndpoint(route *RouteRule) error {
	validRoute, err := route.validate()
	if err != nil {
		return fmt.Errorf("proxy: error handling endpoint: %s", err.Error())
	}
	log.Printf("\tAdding route %s -> %s", validRoute.Path, validRoute.EndpointURL.String())
	handler.routes = append(handler.routes, validRoute)
	return nil
}

func (handler *ProxyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, routeMap := range handler.routes {
		if strings.HasPrefix(request.URL.Path, routeMap.Path) {
			switch routeMap.EndpointURL.Scheme {
			case "ws":
				handler.handleWebsocketRequest(routeMap.EndpointURL, writer, request)
			case "http":
				handler.handleHTTPRequest(routeMap.EndpointURL, writer, request)
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
