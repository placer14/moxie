package proxyhandler

import (
	"strings"
	"testing"
)

func TestNewRoute(t *testing.T) {
	path := "/"
	endpoint := "http://hostname/"

	route, err := newRoute(path, endpoint)
	if err != nil {
		t.Fatal("error creating new route")
	}

	if route.path != path {
		t.Error("path is not set")
	}
	if route.endpointURL.String() != endpoint {
		t.Error("endpoint is not set")
	}
}

func TestNewValidatesPathExists(t *testing.T) {
	emptyPath := ""
	endpoint := "http://hostname/"
	expectedError := "path is empty"

	_, err := newRoute(emptyPath, endpoint)
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestNewValidatesHostExists(t *testing.T) {
	expectedError := "host is empty"

	_, err := newRoute("/", "//")
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestNewValidatesEndpointUrl(t *testing.T) {
	path := "/"
	invalidEndpoint := "http://invalid%23hostname/"
	expectedError := "invalid endpoint url:"

	_, err := newRoute(path, invalidEndpoint)
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}
