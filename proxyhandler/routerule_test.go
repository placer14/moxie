package proxyhandler

import (
	"strings"
	"testing"
)

func TestRouteRuleValidateReturnsValidRouteRule(t *testing.T) {
	route := RouteRule{
		Path:     "/",
		Endpoint: "ws://hostname",
	}

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
	if validRoute.EndpointURL.String() != route.Endpoint {
		t.Errorf("endpointURL is not set\nexpected: %v\nreceived: %v", route.Endpoint, validRoute.EndpointURL.String())
	}
}

func TestValidatesScheme(t *testing.T) {
	expectedScheme := "http"
	route := RouteRule{
		Path:     "/",
		Endpoint: expectedScheme + "://hostname",
	}
	validRoute, err := route.validate()
	if err != nil {
		t.Fatal("expected valid route to be returned")
	}
	if validRoute.EndpointURL.Scheme != expectedScheme {
		t.Errorf("unexpected scheme\nexpected: %v\nreceived: %v", expectedScheme, validRoute.EndpointURL.Scheme)
	}

	expectedScheme = "ws"
	route = RouteRule{
		Path:     "/",
		Endpoint: expectedScheme + "://hostname",
	}
	validRoute, err = route.validate()
	if err != nil {
		t.Fatal("expected valid route to be returned")
	}
	if validRoute.EndpointURL.Scheme != expectedScheme {
		t.Errorf("unexpected scheme\nexpected: %v\nreceived: %v", expectedScheme, validRoute.EndpointURL.Scheme)
	}

	expectedError := "unsupported scheme"
	invalidScheme := "invalid"
	route = RouteRule{
		Path:     "/",
		Endpoint: invalidScheme + "://hostname",
	}
	validRoute, err = route.validate()
	if err == nil {
		t.Fatal("expected route to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidateChecksSchemeIsNotEmpty(t *testing.T) {
	expectedError := "protocol scheme is empty"
	route := RouteRule{
		Path:     "/",
		Endpoint: "//hostname",
	}
	_, err := route.validate()
	if err == nil {
		t.Fatal("expected route to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidateChecksPathIsNotEmpty(t *testing.T) {
	emptyPath := ""
	expectedError := "path is empty"
	route := RouteRule{
		Path:     emptyPath,
		Endpoint: "http://hostname",
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
	emptyHost := "http://"
	route := RouteRule{
		Path:     "/",
		Endpoint: emptyHost,
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
		Path:     "/",
		Endpoint: invalidEndpoint,
	}
	_, err := r.validate()
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}
