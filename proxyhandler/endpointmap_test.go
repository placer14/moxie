package proxyhandler

import (
	"strings"
	"testing"
)

func TestNewCreatesRouteMap(t *testing.T) {
	path := "/"
	endpoint := "http://hostname/"

	endpointMap, err := newEndpointMap(path, endpoint)
	if err != nil {
		t.Fatal("error creating new endpointMap")
	}

	if endpointMap.path != path {
		t.Error("path is not set")
	}
	if endpointMap.endpointURL.String() != endpoint {
		t.Error("endpoint is not set")
	}
}

func TestNewValidatesPathExists(t *testing.T) {
	emptyPath := ""
	endpoint := "http://hostname/"
	expectedError := "path is empty"

	_, err := newEndpointMap(emptyPath, endpoint)
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestNewValidatesHostExists(t *testing.T) {
	expectedError := "host is empty"

	_, err := newEndpointMap("/", "//")
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

	_, err := newEndpointMap(path, invalidEndpoint)
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}
