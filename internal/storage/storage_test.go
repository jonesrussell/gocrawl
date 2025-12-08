package storage_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	loggermocks "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	mockLogger := loggermocks.NewMockInterface(ctrl)
	mockLogger.EXPECT().Error("Index not found", "index", "non_existent_index").Return()

	result, err := storage.NewStorage(storage.StorageParams{
		Client: mockClient,
		Logger: mockLogger,
	})
	require.NoError(t, err)
	s := result.Storage

	// Test searching a non-existent index
	_, err = s.Search(context.Background(), "non_existent_index", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrIndexNotFound)
	require.Contains(t, err.Error(), "non_existent_index")
}

func TestSearch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	mockLogger := loggermocks.NewMockInterface(ctrl)
	result, err := storage.NewStorage(storage.StorageParams{
		Client: mockClient,
		Logger: mockLogger,
	})
	require.NoError(t, err)
	s := result.Storage

	results, err := s.Search(context.Background(), "test-index", map[string]any{
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, results)
	require.Len(t, results, 1)
}

func TestNewStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := loggermocks.NewMockInterface(ctrl)
	transport := &mockTransport{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
		},
	}

	mockClient, err := setupMockClient(transport)
	require.NoError(t, err)

	result, err := storage.NewStorage(storage.StorageParams{
		Client: mockClient,
		Logger: mockLogger,
	})
	require.NoError(t, err)
	store := result.Storage
	assert.NotNil(t, store)
	assert.Implements(t, (*types.Interface)(nil), store)
}

func TestStorage_TestConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	mockLogger := loggermocks.NewMockInterface(ctrl)
	result, err := storage.NewStorage(storage.StorageParams{
		Client: mockClient,
		Logger: mockLogger,
	})
	require.NoError(t, err)
	s := result.Storage

	err = s.TestConnection(context.Background())
	require.NoError(t, err)
}
