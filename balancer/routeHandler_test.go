package balancer

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRouting(t *testing.T) {
	// Given
	ts := stubServer()
	defer ts.Close()

	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Cookie", "someKey=randomValue;")

	w := httptest.NewRecorder()

	requestUrl, _ := new(url.URL).Parse(ts.URL + "/some-address")
	handler := ReverseProxyHandler{
		TargetAddress: *requestUrl,
	}

	// When
	handler.Handle(w, req)

	// Then
	assert.Equal(t, stubHeader.Get("Content-type"), "application/json")
	assert.Equal(t, stubHeader.Get("Cookie"), "someKey=randomValue;")

	assert.Equal(t, stubUri.String(), "/some-address/foo")
}

var stubHeader *http.Header = nil
var stubUri *url.URL = nil

func stubServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stubHeader = &r.Header
		stubUri = r.URL
		_, _ = fmt.Fprintln(w, "Hello, client")
	}))
	return ts
}
