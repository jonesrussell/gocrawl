package storage

import (
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// MockTransport implements http.RoundTripper for testing
type MockTransport struct {
	Response    string
	StatusCode  int
	Error       error
	RequestFunc func(*http.Request) (*http.Response, error)
}

// RoundTrip implements http.RoundTripper
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
func (t *MockTransport) Perform(req *http.Request) (*http.Response, error) {
	return t.RoundTrip(req)
}

// createESResponse creates an Elasticsearch API response from the mock data
func (t *MockTransport) createESResponse() *esapi.Response {
	return &esapi.Response{
		StatusCode: t.StatusCode,
		Body:       io.NopCloser(strings.NewReader(t.Response)),
		Header:     make(http.Header),
	}
}
