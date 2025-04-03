package storage_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
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

// setupMockClient creates a new Elasticsearch client with mock transport
func setupMockClient(transport http.RoundTripper) (*es.Client, error) {
	return es.NewClient(es.Config{
		Transport: transport,
	})
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
	mockClient, err := setupMockClient(transport)
	require.NoError(t, err)

	mockLogger := testutils.NewMockLogger()
	mockLogger.On("Error", "Index not found", []any{"index", "non_existent_index"}).Return()

	s := storage.NewStorage(mockClient, mockLogger, storage.Options{
		IndexName: "test-index",
	})

	// Test searching a non-existent index
	_, err = s.Search(t.Context(), "non_existent_index", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrIndexNotFound)
	require.Contains(t, err.Error(), "non_existent_index")

	// Verify that all expected logger calls were made
	mockLogger.AssertExpectations(t)
}

func TestSearch_Success(t *testing.T) {
	// Create a mock transport that returns a successful search response
	transport := &mockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/_cluster/health" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"green"}`)),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}
			if req.URL.Path == "/test-index/_exists" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"hits": {
						"total": {"value": 1},
						"hits": [{"_source": {"title": "Test Document"}}]
					}
				}`)),
				Header: http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
			}, nil
		},
	}

	mockClient, err := setupMockClient(transport)
	require.NoError(t, err)

	mockLogger := testutils.NewMockLogger()
	s := storage.NewStorage(mockClient, mockLogger, storage.Options{
		IndexName: "test-index",
	})

	results, err := s.Search(t.Context(), "test-index", map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, results)
	require.Len(t, results, 1)
}

func TestNewStorage(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	transport := &mockTransport{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		},
	}

	mockClient, err := setupMockClient(transport)
	require.NoError(t, err)

	opts := storage.Options{
		IndexName: "test-index",
	}

	store := storage.NewStorage(mockClient, mockLogger, opts)
	assert.NotNil(t, store)
	assert.Implements(t, (*types.Interface)(nil), store)
}

func TestStorage_TestConnection(t *testing.T) {
	transport := &mockTransport{
		RoundTripFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"green"}`)),
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
			}, nil
		},
	}

	mockClient, err := setupMockClient(transport)
	require.NoError(t, err)

	mockLogger := testutils.NewMockLogger()
	s := storage.NewStorage(mockClient, mockLogger, storage.Options{
		IndexName: "test-index",
	})

	err = s.TestConnection(t.Context())
	require.NoError(t, err)
}
