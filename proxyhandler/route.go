package proxyhandler

import (
	"errors"
	"fmt"
	"net/url"
)

type route struct {
	path        string
	endpointURL *url.URL
}

func newRoute(path, endpoint string) (*route, error) {
	if len(path) == 0 {
		return nil, errors.New("path is empty")
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %s", err.Error())
	}
	if url.Host == "" {
		return nil, errors.New("host is empty")
	}
	return &route{path: path, endpointURL: url}, nil
}
