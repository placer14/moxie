package proxyhandler_test

import (
	"github.com/jarcoal/httpmock"
	handler "github.com/placer14/proxyhandler"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestBodyTransfer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedBody := "This is the expected body"
	httpmock.RegisterResponder("GET", "http://hostname/", httpmock.NewStringResponder(200, expectedBody))
	req := httptest.NewRequest("GET", "http://hostname/", nil)
	recorder := httptest.NewRecorder()

	h, _ := handler.New("")
	h.ServeHTTP(recorder, req)
	actualBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal("Response body unreadable:", err.Error())
	}

	if expectedBody != string(actualBody) {
		t.Error("Expected body not found")
		t.Log("Expected:", expectedBody)
		t.Log("Actual:", string(actualBody))
	}
}

func TestRequestHeaderTransfer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedHeader := http.Header{
		"X-Foo": []string{"IMPORTANT"},
		"X-Bar": []string{"here; are_some; headers"},
	}
	httpmock.RegisterResponder("GET", "http://hostname/", func(r *http.Request) (*http.Response, error) {
		if !reflect.DeepEqual(r.Header, expectedHeader) {
			t.Error("Unexpected headers")
			t.Log("Expected:", expectedHeader)
			t.Log("Actual:", r.Header)
		}
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("GET", "http://hostname/", nil)
	req.Header = expectedHeader

	h, _ := handler.New("")
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestResponseHeaderTransfer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

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
	httpmock.RegisterResponder("GET", "http://hostname/", mockResponder)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://hostname/", nil)

	h, _ := handler.New("")
	h.ServeHTTP(recorder, req)

	if !reflect.DeepEqual(recorder.Header(), expectedHeader) {
		t.Error("Unexpected headers")
		t.Log("Expected:", expectedHeader)
		t.Log("Actual:", recorder.Header())
	}
}

func TestPostMethod(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedPostBody := `{"some":"json"}`
	success := false
	httpmock.RegisterResponder("POST", "http://hostname/", func(r *http.Request) (*http.Response, error) {
		success = true
		actualBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal("Response body unreadable:", err.Error())
		}
		if expectedPostBody != string(actualBody) {
			t.Error("Body did not match")
			t.Log("Expected:", expectedPostBody)
			t.Log("Actual:", string(actualBody))
		}
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("POST", "http://hostname/", strings.NewReader(expectedPostBody))
	recorder := httptest.NewRecorder()

	h, _ := handler.New("")
	h.ServeHTTP(recorder, req)
	if !success {
		t.Error("Expected POST responder to be executed")
	}
}

func TestProxiedRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	success := false
	httpmock.RegisterResponder("GET", "http://google.com/foo", func(r *http.Request) (*http.Response, error) {
		success = true
		return httpmock.NewStringResponse(200, ""), nil
	})
	req := httptest.NewRequest("GET", "http://hostname/foo", nil)
	recorder := httptest.NewRecorder()

	h, _ := handler.New("")
	overrideMask, _ := url.Parse("//google.com/")
	h.HandleEndpoint("/foo", overrideMask)
	h.ServeHTTP(recorder, req)

	if !success {
		t.Error("Expected handler to direct request to google.com host")
	}
}

func TestDefaultHostParsingFailure(t *testing.T) {
	_, err := handler.New("http://192.168%31/") // invalid URL
	if err == nil {
		t.Fatal("Expected handler to return an error when defaultProxiedServer cannot be parsed")
	}
}

func TestDefaultHostIsSet(t *testing.T) {
	defaultHost := "http://192.168.1.1"
	h, _ := handler.New(defaultHost) // invalid URL
	if h.DefaultHost.String() != defaultHost {
		t.Fatal("Expected handler's DefaultHost to match argument")
	}
}

func TestDefaultHostIsUsedWhenMatchingRouteMissing(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	success := false

	httpmock.RegisterResponder("GET", "//hostname/", func(r *http.Request) (*http.Response, error) {
		success = true
		return httpmock.NewStringResponse(200, ""), nil
	})

	h, _ := handler.New("//hostname")
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !success {
		t.Error("Expected default host to be requested")
	}
}

func BenchmarkRouteHandling(b *testing.B) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	h, _ := handler.New("//defaultRoute/")
	proxyEndpoints := map[string]string{
		"/":       "//reddit.com/",
		"/foo":    "//cnn.com",
		"/bazqux": "//elsewhere.com",
	}

	for endpoint, override := range proxyEndpoints {
		u, _ := url.Parse(override)
		h.HandleEndpoint(endpoint, u)

		httpmock.RegisterResponder("GET", override, func(r *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, ""), nil
		})
	}

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	}
}
