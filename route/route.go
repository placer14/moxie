package route

import (
	"fmt"
	"net/url"
)

type Route struct {
	Path        string
	EndpointURL *url.URL
}

func NewRoute(path, endpoint string) (*Route, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path is empty")
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %s", err.Error())
	}
	if url.Host == "" {
		return nil, fmt.Errorf("host is empty")
	}
	return &Route{Path: path, EndpointURL: url}, nil
}
