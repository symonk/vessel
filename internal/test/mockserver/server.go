package mockserver

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
)

// ServerOption defines a functional option for the MockServer
// that registers custom handlers for testing.
type ServerOption func(*MockServer)

// MockServer wraps an httptest server offering a fluent api
// for various different built in handlers to simplify and make
// unit testing easier.
type MockServer struct {
	Server *httptest.Server
	mux    *http.ServeMux
	Seen   atomic.Int64
}

// New instantiates a new MockServer and returns a ptr to it.
// Functional options are applied and the caller is responsible
// for calling .Close() when finished with the server.
func New(options ...ServerOption) *MockServer {
	mux := http.NewServeMux()
	s := httptest.NewServer(mux)
	m := &MockServer{
		Server: s,
		mux:    mux,
	}
	for _, opt := range options {
		opt(m)
	}
	return m
}

// Close closes the server cleanly
func (m *MockServer) Close() {
	if m != nil && m.Server != nil {
		m.Server.Close()
	}
}

// WithStatusCodeTestHandler registers a handler listening on
// /status/$code to return an empty response body with that
// particular stats code.
func WithStatusCodeTestHandler() ServerOption {
	return func(m *MockServer) {
		m.mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/status/"), "/")
			if len(parts) < 1 {
				http.Error(w, "missing status code", http.StatusBadRequest)
				return
			}

			code, err := strconv.Atoi(parts[0])
			if err != nil || code < 100 || code > 599 {
				http.Error(w, "invalid status code", http.StatusBadRequest)
				return
			}
			m.Seen.Add(1)
			w.WriteHeader(code)
		})
	}
}
