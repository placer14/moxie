package proxyhandler

import (
	"fmt"
	"net/url"
)

// Configuration controls the behavior of a newly created ProxyHandler.
// DefaultRoute is a URL string which is the target of any requests
// that are not matched to any RouteRules in Routes. Each inbound request
// has its URL.Path matched against each of the RouteRule.Path in the order
// listed. The RouteRule.Path will match if it has the prefix of the
// request URL.Path.
type Configuration struct {
	DefaultRoute string
	Routes       []*RouteRule
}

type validConfiguration struct {
	DefaultRoute *url.URL
	Routes       []*validRouteRule
}

func (config *Configuration) validate() (*validConfiguration, error) {
	var err error
	var validConfig = &validConfiguration{}
	if len(config.DefaultRoute) == 0 {
		return nil, fmt.Errorf("default route is missing")
	}
	validConfig.DefaultRoute, err = url.Parse(config.DefaultRoute)
	if err != nil {
		return nil, fmt.Errorf("invalid default route: %s", err.Error())
	}
	if len(config.Routes) == 0 {
		return nil, fmt.Errorf("no configured routes")
	}
	validConfig.Routes = make([]*validRouteRule, len(config.Routes))
	for index, route := range config.Routes {
		validRoute, err := route.validate()
		if err != nil {
			return nil, fmt.Errorf("invalid RouteRule: %s", err.Error())
		}
		validConfig.Routes[index] = validRoute
	}
	return validConfig, nil
}
