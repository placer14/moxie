package proxyhandler

import (
	"strings"
	"testing"
)

func TestRouteRuleValidateReturnsValidRouteRule(t *testing.T) {
	beforeTest()
	defer afterTest()

	route := RouteRule{
		Path:             "/",
		Endpoint:         "//hostname",
		WebsocketEnabled: true,
	}

	validRoute, err := route.validate()
	if err != nil {
		t.Fatal("error validating route")
	}

	if validRoute.Path != route.Path {
		t.Error("path is not set")
	}
	if validRoute.Endpoint != route.Endpoint {
		t.Error("endpoint is not set")
	}
	if validRoute.EndpointURL.String() != route.Endpoint {
		t.Error("endpointURL is not set")
	}
	if validRoute.WebsocketEnabled != route.WebsocketEnabled {
		t.Error("websocketEnabled is not set")
	}
}

func TestValidateChecksPathIsNotEmpty(t *testing.T) {
	beforeTest()
	defer afterTest()

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
	beforeTest()
	defer afterTest()

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
