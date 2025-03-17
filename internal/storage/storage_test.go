package storage_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/storage"
	testutils "github.com/jonesrussell/gocrawl/internal/storage/testing"
	"github.com/stretchr/testify/require"
)

// mockTransport implements http.RoundTripper for mocking Elasticsearch responses
type mockTransport struct {
	Response    *http.Response
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripFn != nil {
		return t.RoundTripFn(req)
	}
	return t.Response, nil
}

func TestSearch_IndexNotFound(t *testing.T) {
	// Create a mock transport that returns 404 for index existence check
	transport := &mockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":{"type":"index_not_found_exception"}}`)),
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
			}, nil
		},
	}

	// Create a client with the mock transport
	mockClient, err := es.NewClient(es.Config{
		Transport: transport,
	})
	require.NoError(t, err)

	s := &storage.ElasticsearchStorage{
		ESClient: mockClient,
		Logger:   testutils.NewNopLogger(),
	}

	// Test searching a non-existent index
	_, err = s.Search(t.Context(), "non_existent_index", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrIndexNotFound)
	require.Contains(t, err.Error(), "non_existent_index")
}
