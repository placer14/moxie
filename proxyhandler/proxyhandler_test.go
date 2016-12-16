package proxyhandler

import (
	"bytes"
	"fmt"
	"github.com/jarcoal/httpmock"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func beforeTest() {
	httpmock.Activate()
	log.SetOutput(ioutil.Discard)
}

func afterTest() {
	log.SetOutput(os.Stderr)
	httpmock.DeactivateAndReset()
}

func TestNewReturnsValidProxyHandler(t *testing.T) {
	config := buildConfiguration()
	_, err := New(config)
	if err != nil {
		t.Fatal("expected valid ProxyHandler to be created")
	}
}

func TestNewReturnsErrorWithInvalidConfiguration(t *testing.T) {
	expectedError := "invalid configuration"
	invalidConfig := buildConfiguration()
	invalidConfig.DefaultRoute = "http://invalid%123.hostname"
	_, err := New(invalidConfig)
	if err == nil {
		t.Fatal("expected invalid configuration to return an error")
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected validConfiguration route to be in returned config\nexpected: %v\nactual: %v",
			expectedError, err.Error(),
		)
	}
}

func TestResponseStatus(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedStatus := 999
	httpmock.RegisterResponder("GET", "http://defaulthost/", httpmock.NewBytesResponder(expectedStatus, nil))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("Failed creating handler: %v", err.Error())
	}
	h.ServeHTTP(recorder, req)

	if recorder.Code != expectedStatus {
		t.Fatalf("Expected status code not found\n\tExpected: %v\n\tActual: %v", expectedStatus, recorder.Code)
	}
}

func TestRequestBodyTransfer(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedBody := []byte("This is the expected body")
	httpmock.RegisterResponder("POST", "http://defaulthost/", func(r *http.Request) (*http.Response, error) {
		actualBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Request body unreadable: %v", err.Error())
		}

		if string(expectedBody) != string(actualBody) {
			t.Fatalf("Expected body not found\n\tExpected: %v\n\tActual: %v", expectedBody, actualBody)
		}
		return httpmock.NewStringResponse(200, ""), nil
	})

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(expectedBody))

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestResponseBodyTransfer(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedBody := []byte("This is the expected body")
	httpmock.RegisterResponder("GET", "http://defaulthost/", httpmock.NewBytesResponder(200, expectedBody))
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(recorder, req)
	actualBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatalf("Response body unreadable: %v", err.Error())
	}

	if string(expectedBody) != string(actualBody) {
		t.Fatalf("Expected body not found\n\tExpected: %v\n\tActual: %v", expectedBody, actualBody)
	}
}

func TestRequestHeaderTransfer(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedHeader := http.Header{
		"X-Foo": []string{"IMPORTANT"},
		"X-Bar": []string{"here; are_some; headers"},
	}
	httpmock.RegisterResponder("GET", "http://defaulthost/", func(r *http.Request) (*http.Response, error) {
		if !reflect.DeepEqual(r.Header, expectedHeader) {
			t.Fatalf("Unexpected headers\n\tExpected: %v\n\tActual: %v", expectedHeader, r.Header)
		}
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header = expectedHeader

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestResponseHeaderTransfer(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedHeader := http.Header{
		"Accept-Ranges":  []string{"bytes"},
		"Content-Length": []string{"6"},
		"Content-Type":   []string{"text/plain; charset=utf-8"},
		"Content-Range":  []string{"bytes 0-5/1862"},
	}
	mockResponder := httpmock.ResponderFromResponse(&http.Response{
		Header: expectedHeader,
		Body:   httpmock.NewRespBodyFromString(""),
	})
	httpmock.RegisterResponder("GET", "http://defaulthost/", mockResponder)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(recorder, req)

	if !reflect.DeepEqual(recorder.Header(), expectedHeader) {
		t.Fatalf("Unexpected headers\n\tExpected: %v\n\tActual: %v", expectedHeader, recorder.Header())
	}
}

func TestPostMethod(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedPostBody := `{"some":"json"}`
	success := false
	httpmock.RegisterResponder("POST", "http://defaulthost/", func(r *http.Request) (*http.Response, error) {
		success = true
		actualBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Response body unreadable: %v", err.Error())
		}
		if expectedPostBody != string(actualBody) {
			t.Fatalf("Body did not match\n\tExpected: %v\n\tActual: %v", expectedPostBody, string(actualBody))
		}
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("POST", "/", strings.NewReader(expectedPostBody))
	recorder := httptest.NewRecorder()

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(recorder, req)
	if !success {
		t.Error("Expected POST responder to be executed")
	}
}

func TestProxyHandlesSpecificEndpoint(t *testing.T) {
	beforeTest()
	defer afterTest()

	success := false
	httpmock.RegisterResponder("GET", "http://anotherhost/foo", func(r *http.Request) (*http.Response, error) {
		success = true
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("GET", "/foo", nil)
	recorder := httptest.NewRecorder()

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	config.Routes = []*RouteRule{
		&RouteRule{Path: "/foo", Endpoint: "http://anotherhost"},
	}
	h, err := New(config)
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	h.ServeHTTP(recorder, req)
	if !success {
		t.Error("Expected handler to direct request to //anotherhost")
	}
}

func TestDefaultHostIsUsedWhenMatchingRouteMissing(t *testing.T) {
	beforeTest()
	defer afterTest()
	success := false

	httpmock.RegisterResponder("GET", "http://notgoogle/", func(r *http.Request) (*http.Response, error) {
		success = true
		return httpmock.NewStringResponse(200, ""), nil
	})

	config := buildConfiguration()
	config.DefaultRoute = "http://notgoogle"
	h, _ := New(config)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !success {
		t.Error("Expected default host to be requested")
	}
}

func BenchmarkRouteHandling(b *testing.B) {
	beforeTest()
	defer afterTest()

	config := buildConfiguration()
	config.DefaultRoute = "http://defaulthost"
	config.Routes = []*RouteRule{
		&RouteRule{Path: "/bazqux", Endpoint: "http://elsewhere.com"},
		&RouteRule{Path: "/foo", Endpoint: "http://cnn.com"},
		&RouteRule{Path: "/", Endpoint: "http://reddit.com"},
	}
	h, err := New(config)
	if err != nil {
		b.Fatal("unable to create proxy")
	}

	endpointRequests := make(map[string]int)
	httpmock.RegisterNoResponder(func(r *http.Request) (*http.Response, error) {
		endpointRequests[r.URL.Host]++
		return httpmock.NewStringResponse(200, ""), nil
	})

	pathRequests := make(map[string]int)
	for i := 0; i < b.N; i++ {
		n := rand.Int() % len(config.Routes)
		pathRequests[config.Routes[n].Path]++
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", config.Routes[n].Path, nil))
	}

	var result = []byte("\nPath     Requests     Hostname        Received Delta\n")
	for _, route := range config.Routes {
		u, _ := url.Parse(route.Endpoint)
		pathHits := pathRequests[route.Path]
		endpointHits := endpointRequests[u.Host]
		result = append(result, fmt.Sprintf("% 8s  % 4v  %20s   % 4v    %04v  \n", route.Path, pathHits, route.Endpoint, endpointHits, pathHits-endpointHits)...)
		if pathHits != endpointHits {
			b.Errorf("Requests made to %s do not match requests received by %s\nExpected: %d\nActual %d\n", route.Path, route.Endpoint, pathHits, endpointHits)
		}
	}
	fmt.Println(string(result))
}
