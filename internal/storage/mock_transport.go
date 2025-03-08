package storage

import (
	"bytes"
	"io"
	"net/http"
)

// MockTransport implements http.RoundTripper for testing
type MockTransport struct {
	Response   string
	StatusCode int
}

// RoundTrip implements http.RoundTripper
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.StatusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.Response)),
		Header:     make(http.Header),
	}, nil
}
