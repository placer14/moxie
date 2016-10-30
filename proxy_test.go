package proxy_handler_test

import (
	"github.com/jarcoal/httpmock"
	handler "github.com/placer14/proxy_handler"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBodyTransfer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedBody := "This is the expected body"
	httpmock.RegisterResponder("GET", "/", httpmock.NewStringResponder(200, expectedBody))
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	h := handler.NewProxyHandler()
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

func TestHeaderTransfer(t *testing.T) {
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
	httpmock.RegisterResponder("GET", "/", mockResponder)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	h := handler.NewProxyHandler()
	h.ServeHTTP(recorder, req)

	if !reflect.DeepEqual(recorder.Header(), expectedHeader) {
		t.Error("Unexpected headers")
		t.Log("Expected: ", expectedHeader)
		t.Log("Actual: ", recorder.Header())
	}
}
