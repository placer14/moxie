package proxyhandler

import (
	"strings"
	"testing"
)

func buildConfiguration() *Configuration {
	return &Configuration{
		DefaultRoute: "http://default.endpoint",
		Routes: []*RouteRule{
			&RouteRule{Path: "/route1", Endpoint: "http://endpoint.one"},
		},
	}
}

func TestValidationProducesValidConfiguration(t *testing.T) {
	config := buildConfiguration()
	validConfig, err := config.validate()
	if err != nil {
		t.Fatal("expected validConfiguration to be returned without error")
	}

	// These checks verify a field `EndpointURL` which only exists on validRouteRule
	// and confirms validate() is returning the appropriate types
	if validConfig.DefaultRoute.String() != config.DefaultRoute {
		t.Errorf("expected validConfiguration DefaultRoute to be in returned config\nexpected: %v\nactual: %v",
			config.DefaultRoute,
			validConfig.DefaultRoute.String(),
		)
	}
	for index, route := range config.Routes {
		if validConfig.Routes[index].EndpointURL.String() != route.Endpoint {
			t.Errorf("expected validConfiguration route to be in returned config\nexpected: %v\nactual: %v",
				route.Endpoint,
				validConfig.Routes[index].EndpointURL.String(),
			)
		}
	}

}

func TestValidationChecksRoutesArePresent(t *testing.T) {
	expectedError := "no configured routes"
	config := buildConfiguration()
	config.Routes = []*RouteRule{}

	_, err := config.validate()
	if err == nil {
		t.Fatal("expected config to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidationDetectsInvalidRouteRules(t *testing.T) {
	expectedError := "invalid RouteRule"
	config := buildConfiguration()
	config.Routes = []*RouteRule{
		&RouteRule{Path: "/foo", Endpoint: "http://good"},
		&RouteRule{Path: "", Endpoint: "http://bad"},
		&RouteRule{Path: "/baz", Endpoint: "http://good"},
	}

	_, err := config.validate()
	if err == nil {
		t.Fatal("expected config to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidationChecksDefaultRouteIsPresent(t *testing.T) {
	expectedError := "default route is missing"
	config := buildConfiguration()
	config.DefaultRoute = ""

	_, err := config.validate()
	if err == nil {
		t.Fatal("expected config to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}

func TestValidationChecksDefaultRouteIsValid(t *testing.T) {
	expectedError := "invalid default route"
	config := buildConfiguration()
	config.DefaultRoute = "http://inv%4al.id"

	_, err := config.validate()
	if err == nil {
		t.Fatal("expected config to be invalid")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error not found\nexpected: %v\nreceived: %v", expectedError, err.Error())
	}
}
