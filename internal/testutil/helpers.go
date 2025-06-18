package testutil

import (
	"net/http"
	"net/http/httptest"
)

// NewMockServer returns a basic http test server.
func NewMockServer(response string, status int) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			w.Write([]byte(response))
		}))
}
