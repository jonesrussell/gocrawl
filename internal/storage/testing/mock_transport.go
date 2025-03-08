package testing

import (
	"bytes"
	"io"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockTransport implements http.RoundTripper for testing
type MockTransport struct {
	mock.Mock
}

// RoundTrip implements http.RoundTripper
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	if resp := args.Get(0); resp != nil {
		return resp.(*http.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateMockResponse creates a mock HTTP response with given status code and body
func CreateMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}
