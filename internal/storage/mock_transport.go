package storage

import (
	"io"
	"net/http"
	"strings"
)

// mockTransport implements http.RoundTripper for testing
type mockTransport struct {
	Response    string
	StatusCode  int
	Error       error
	RequestFunc func(*http.Request) (*http.Response, error)
}

// RoundTrip implements http.RoundTripper
func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Error != nil {
		return nil, t.Error
	}

	if t.RequestFunc != nil {
		return t.RequestFunc(req)
	}

	return &http.Response{
		StatusCode: t.StatusCode,
		Body:       io.NopCloser(strings.NewReader(t.Response)),
		Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
	}, nil
}

// Perform implements the Transport interface
func (t *mockTransport) Perform(req *http.Request) (*http.Response, error) {
	return t.RoundTrip(req)
}
