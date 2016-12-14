package proxyhandler

import (
	"fmt"
	"net/url"
)

// RouteRule represents a route which the proxyHandler can use to direct requests to
// appropriate backend system. Path is the requested path in the URL received by the
// proxyHandler. Endpoint is the backend host to direct the traffic to. WebsocketEnabled
// instructs the proxyHandler to attempt to upgrade these connections to websockets
// before establishing the connection to the Endpoint.
type RouteRule struct {
	Path             string
	Endpoint         string
	WebsocketEnabled bool
}

type validRouteRule struct {
	RouteRule
	EndpointURL *url.URL
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
	validRoute := validRouteRule{
		RouteRule: RouteRule{
			Path:             route.Path,
			Endpoint:         route.Endpoint,
			WebsocketEnabled: route.WebsocketEnabled,
		},
		EndpointURL: endpointURL,
	}
	return &validRoute, nil
}
