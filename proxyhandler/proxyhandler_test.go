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

func TestResponseStatus(t *testing.T) {
	beforeTest()
	defer afterTest()

	expectedStatus := 999
	httpmock.RegisterResponder("GET", "http://defaulthost/", httpmock.NewBytesResponder(expectedStatus, nil))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
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

	h, err := New("http://defaulthost")
	if err != nil {
		t.Fatalf("unable to create proxyhandler: %s", err.Error())
	}
	route := &RouteRule{
		Path:     "/foo",
		Endpoint: "http://anotherhost",
	}
	h.HandleEndpoint(route)
	h.ServeHTTP(recorder, req)
	if !success {
		t.Error("Expected handler to direct request to //anotherhost")
	}
}

func TestDefaultHostParsingFailure(t *testing.T) {
	expectedError := "proxy: invalid default host"
	_, err := New("http://192.168%31/") // invalid URL
	if err == nil {
		t.Fatal("Expected handler to return an error when default host cannot be parsed")
	}
	if !strings.HasPrefix(err.Error(), expectedError) {
		t.Errorf("Expected invalid default host error\nActual: %v\nExpected: %v", err.Error(), expectedError)
	}
}

func TestDefaultHostContainsHost(t *testing.T) {
	defaultHost := "http://"
	_, err := New(defaultHost) // invalid URL
	if err == nil {
		t.Fatal("Expected handler to return an error when default host is empty")
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

	h, _ := New("http://notgoogle")
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !success {
		t.Error("Expected default host to be requested")
	}
}

func TestHandleEndpointReturnsError(t *testing.T) {
	beforeTest()
	defer afterTest()

	invalidRoute := &RouteRule{
		Path:     "", // empty path makes route invalid
		Endpoint: "http://anotherhostname",
	}
	expectedError := "error handling endpoint"
	h, _ := New("http://defaulthost")

	actualError := h.HandleEndpoint(invalidRoute)
	if actualError == nil {
		t.Fatal("expected invalid route to return error")
	}
	if !strings.Contains(actualError.Error(), expectedError) {
		t.Errorf("Expected handle endpoint error\nActual: %v\nExpected: %v", actualError.Error(), expectedError)
	}
}

func BenchmarkRouteHandling(b *testing.B) {
	beforeTest()
	defer afterTest()

	h, err := New("http://defaulthost")
	if err != nil {
		b.Fatal("unable to create proxy")
	}
	routes := []*RouteRule{
		&RouteRule{Path: "/bazqux", Endpoint: "http://elsewhere.com"},
		&RouteRule{Path: "/foo", Endpoint: "http://cnn.com"},
		&RouteRule{Path: "/", Endpoint: "http://reddit.com"},
	}
	for _, r := range routes {
		err = h.HandleEndpoint(r)
		if err != nil {
			b.Fatalf("unable to handle endpoint: %v -> %v", r.Path, r.Endpoint)
		}
	}

	endpointRequests := make(map[string]int)
	httpmock.RegisterNoResponder(func(r *http.Request) (*http.Response, error) {
		endpointRequests[r.URL.Host]++
		return httpmock.NewStringResponse(200, ""), nil
	})

	pathRequests := make(map[string]int)
	for i := 0; i < b.N; i++ {
		n := rand.Int() % len(routes)
		pathRequests[routes[n].Path]++
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", routes[n].Path, nil))
	}

	var result = []byte("\nPath     Requests     Hostname        Received Delta\n")
	for _, route := range routes {
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
