package proxyhandler

import (
	"errors"
	"net/url"
)

type endpointMap struct {
	path        string
	endpointURL *url.URL
}

func newEndpointMap(path, endpoint string) (route *endpointMap, err error) {
	if len(path) == 0 {
		return nil, errors.New("path is empty")
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.New("invalid endpoint url: " + err.Error())
	}
	if url.Host == "" {
		return nil, errors.New("host is empty")
	}
	route = &endpointMap{
		path:        path,
		endpointURL: url,
	}
	return
}
