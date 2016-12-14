package proxyhandler

import (
	"strings"
	"testing"
)

func TestRouteRuleValidateReturnsValidRouteRule(t *testing.T) {
	route := RouteRule{
		Path:             "/",
		Endpoint:         "//hostname",
		WebsocketEnabled: true,
	}
	expectedEndpointURL := "ws://hostname"

	validRoute, err := route.validate()
	if err != nil {
		t.Fatal("error validating route")
	}

	if validRoute.Path != route.Path {
		t.Errorf("path is not set\nexpected: %v\nreceived: %v", route.Path, validRoute.Path)
	}
	if validRoute.Endpoint != route.Endpoint {
		t.Errorf("endpoint is not set\nexpected: %v\nreceived: %v", route.Endpoint, validRoute.Endpoint)
	}
	if validRoute.EndpointURL.String() != expectedEndpointURL {
		t.Errorf("endpointURL is not set\nexpected: %v\nreceived: %v", expectedEndpointURL, validRoute.EndpointURL.String())
	}
	if validRoute.WebsocketEnabled != route.WebsocketEnabled {
		t.Errorf("endpointURL is not set\nexpected: %v\nreceived: %v", route.WebsocketEnabled, validRoute.WebsocketEnabled)
		t.Error("websocketEnabled is not set")
	}
}

func TestValidateSetsDefaultHTTPScheme(t *testing.T) {
	expectedScheme := "http"
	route := RouteRule{
		Path:             "/",
		Endpoint:         "//hostname",
		WebsocketEnabled: false,
	}
	validRoute, err := route.validate()
	if err != nil {
		t.Fatal("expected valid route to be returned")
	}
	if validRoute.EndpointURL.Scheme != expectedScheme {
		t.Errorf("unexpected non-websocket rule scheme\nexpected: %v\nreceived: %v", expectedScheme, validRoute.EndpointURL.Scheme)
	}
}

func TestValidateSetsDefaultWebsocketScheme(t *testing.T) {
	expectedScheme := "ws"
	route := RouteRule{
		Path:             "/",
		Endpoint:         "//hostname",
		WebsocketEnabled: true,
	}
	validRoute, err := route.validate()
	if err != nil {
		t.Fatal("expected valid route to be returned")
	}
	if validRoute.EndpointURL.Scheme != expectedScheme {
		t.Errorf("unexpected non-websocket rule scheme\nexpected: %v\nreceived: %v", expectedScheme, validRoute.EndpointURL.Scheme)
	}
}

func TestValidateChecksPathIsNotEmpty(t *testing.T) {
	emptyPath := ""
	expectedError := "path is empty"
	route := RouteRule{
		Path:             emptyPath,
		Endpoint:         "//hostname",
		WebsocketEnabled: false,
	}

	_, err := route.validate()
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidateChecksHostIsNotEmpty(t *testing.T) {
	expectedError := "host is empty"
	emptyHost := "//"
	route := RouteRule{
		Path:             "/",
		Endpoint:         emptyHost,
		WebsocketEnabled: false,
	}

	_, err := route.validate()
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidateVerifiesEndpoint(t *testing.T) {
	invalidEndpoint := "http://invalid%23hostname/"
	expectedError := "parsing endpoint"

	r := RouteRule{
		Path:             "/",
		Endpoint:         invalidEndpoint,
		WebsocketEnabled: false,
	}
	_, err := r.validate()
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}
