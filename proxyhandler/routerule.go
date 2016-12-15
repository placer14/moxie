package proxyhandler

import (
	"fmt"
	"net/url"
)

// RouteRule represents a route which the proxyHandler can use to direct requests to
// appropriate backend system. Path is the requested path in the URL received by the
// proxyHandler. Endpoint is the backend host to direct the traffic to.
type RouteRule struct {
	Path     string
	Endpoint string
}

type validRouteRule struct {
	RouteRule
	EndpointURL *url.URL
}

var validSchemes = map[string]struct{}{
	"ws":   struct{}{},
	"http": struct{}{},
}

func (route RouteRule) validate() (*validRouteRule, error) {
	if len(route.Path) == 0 {
		return nil, fmt.Errorf("path is empty")
	}
	endpointURL, err := url.Parse(route.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint: %s", err.Error())
	}
	if len(endpointURL.Host) == 0 {
		return nil, fmt.Errorf("host is empty")
	}
	if endpointURL.Scheme == "" {
		return nil, fmt.Errorf("protocol scheme is empty")
	}
	if _, ok := validSchemes[endpointURL.Scheme]; !ok {
		return nil, fmt.Errorf("unsupported scheme: %s", endpointURL.Scheme)
	}
	validRoute := validRouteRule{
		RouteRule: RouteRule{
			Path:     route.Path,
			Endpoint: route.Endpoint,
		},
		EndpointURL: endpointURL,
	}
	return &validRoute, nil
}
