package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupTestStorage creates a new storage instance for testing
func setupTestStorage(t *testing.T) Storage {
	t.Helper()

	// Create test logger
	log := logger.NewMockCustomLogger()

	// Create test config
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: getTestElasticURL(),
		},
		Crawler: config.CrawlerConfig{
			Transport: http.DefaultTransport,
		},
	}

	// Create storage instance
	storage, err := NewStorage(cfg, log)
	require.NoError(t, err, "Failed to create test storage")
	require.NotNil(t, storage.Storage, "Storage instance should not be nil")

	return storage.Storage
}

// getTestElasticURL returns the Elasticsearch URL for testing
func getTestElasticURL() string {
	if url := os.Getenv("TEST_ELASTICSEARCH_URL"); url != "" {
		return url
	}
	return "http://localhost:9200"
}

// CleanupTestIndex removes the test index after tests
func CleanupTestIndex(ctx context.Context, t *testing.T, s Storage, indexName string) {
	t.Helper()

	err := s.DeleteIndex(ctx, indexName)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test index %s: %v", indexName, err)
	}
}

func TestNewStorageWithClient(t *testing.T) {
	// Create test logger
	log := logger.NewMockCustomLogger()

	// Create test config
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
		Crawler: config.CrawlerConfig{
			Transport: http.DefaultTransport,
		},
	}

	// Test storage creation
	result, err := NewStorage(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, result.Storage)
}

func TestStorage_TestConnection(t *testing.T) {
	// Create test logger
	log := logger.NewMockCustomLogger()

	// Create test config
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
		Crawler: config.CrawlerConfig{
			Transport: http.DefaultTransport,
		},
	}

	// Create storage instance
	result, err := NewStorage(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, result.Storage)

	// Test connection
	err = result.Storage.TestConnection(context.Background())
	assert.NoError(t, err)
}

func TestElasticsearchStorage_Operations(t *testing.T) {
	ctx := context.Background()
	log := logger.NewMockCustomLogger()

	// Set up expectations for the logger
	log.On("Info", mock.Anything, mock.Anything).Return()

	// Mock successful response with proper Elasticsearch format
	successResponse := `{
		"took": 1,
		"errors": false,
		"_shards": {
			"total": 2,
			"successful": 2,
			"failed": 0
		},
		"hits": {
			"total": {"value": 1, "relation": "eq"},
			"hits": [{
				"_index": "test-index",
				"_id": "test-id",
				"_score": 1.0,
				"_source": {
					"title": "Test Document",
					"body": "Test Content"
				}
			}]
		}
	}`

	// Define mockTransport here
	mockTransport := &MockTransport{
		Response:   successResponse,
		StatusCode: http.StatusOK,
	}

	// Create elasticsearch client with mock transport
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
		Addresses: []string{"http://localhost:9200"},
	})
	require.NoError(t, err)

	// Create storage with mocked client
	storage := &ElasticsearchStorage{
		ESClient: esClient,
		Logger:   log,
		opts:     DefaultOptions(),
	}

	t.Run("IndexDocument", func(t *testing.T) {
		// Update transport response for index operation
		mockTransport.Response = `{
			"_index": "test-index",
			"_id": "test-id",
			"_version": 1,
			"result": "created",
			"_shards": {
				"total": 2,
				"successful": 2,
				"failed": 0
			},
			"_seq_no": 0,
			"_primary_term": 1
		}`

		doc := map[string]interface{}{
			"title": "Test Document",
			"body":  "Test Content",
		}
		err := storage.IndexDocument(ctx, "test-index", "test-id", doc)
		assert.NoError(t, err)
	})

	t.Run("Search", func(t *testing.T) {
		// Update transport response for search operation
		mockTransport.Response = successResponse

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := storage.Search(ctx, "test-index", query)
		assert.NoError(t, err)
		assert.NotNil(t, results)
	})

	t.Run("BulkIndex", func(t *testing.T) {
		// Update transport response for bulk operation
		mockTransport.Response = `{
			"took": 30,
			"errors": false,
			"items": [
				{
					"index": {
						"_index": "test-index",
						"_id": "1",
						"_version": 1,
						"result": "created",
						"status": 201
					}
				},
				{
					"index": {
						"_index": "test-index",
						"_id": "2",
						"_version": 1,
						"result": "created",
						"status": 201
					}
				}
			]
		}`

		docs := []interface{}{
			map[string]interface{}{"title": "Doc 1"},
			map[string]interface{}{"title": "Doc 2"},
		}
		err := storage.BulkIndex(ctx, "test-index", docs)
		assert.NoError(t, err)

		// Test error case
		mockTransport.Error = fmt.Errorf("bulk index error")
		err = storage.BulkIndex(ctx, "test-index", docs)
		assert.Error(t, err)
		mockTransport.Error = nil
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		update := map[string]interface{}{
			"title": "Updated Title",
		}
		err := storage.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.NoError(t, err)
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := storage.DeleteDocument(ctx, "test-index", "test-id")
		assert.NoError(t, err)
	})

	t.Run("ScrollSearch", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resultChan, err := storage.ScrollSearch(ctx, "test-index", query, 100)
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

		// Read some results
		for result := range resultChan {
			assert.NotNil(t, result)
		}
	})

	t.Run("TestConnection", func(t *testing.T) {
		mockTransport.Response = `{
			"name" : "node-1",
			"cluster_name" : "elasticsearch",
			"version" : {
				"number" : "8.0.0"
			}
		}`
		err := storage.TestConnection(ctx)
		assert.NoError(t, err)

		// Test error case
		mockTransport.Error = fmt.Errorf("connection error")
		err = storage.TestConnection(ctx)
		assert.Error(t, err)
		mockTransport.Error = nil
	})

	t.Run("CreateIndex", func(t *testing.T) {
		mockTransport.Response = `{
			"acknowledged": true,
			"shards_acknowledged": true,
			"index": "test-index"
		}`

		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
				},
			},
		}
		err := storage.CreateIndex(ctx, "test-index", mapping)
		assert.NoError(t, err)

		// Test error case
		mockTransport.Error = fmt.Errorf("create index error")
		err = storage.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)
		mockTransport.Error = nil
	})

	t.Run("DeleteIndex", func(t *testing.T) {
		mockTransport.Response = `{
			"acknowledged": true
		}`
		err := storage.DeleteIndex(ctx, "test-index")
		assert.NoError(t, err)

		// Test error case
		mockTransport.Error = fmt.Errorf("delete index error")
		err = storage.DeleteIndex(ctx, "test-index")
		assert.Error(t, err)
		mockTransport.Error = nil
	})

	t.Run("ScrollSearch_Error", func(t *testing.T) {
		// Set error response for initial search
		mockTransport.Response = `{
			"error": {
				"type": "search_phase_execution_exception",
				"reason": "scroll error"
			},
			"status": 500
		}`
		mockTransport.StatusCode = http.StatusInternalServerError

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resultChan, err := storage.ScrollSearch(ctx, "test-index", query, 100)
		assert.Error(t, err)
		assert.Nil(t, resultChan)

		// Reset transport for subsequent tests
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("ProcessHits_InvalidHit", func(t *testing.T) {
		mockTransport.Response = `{
			"took": 1,
			"hits": {
				"hits": [
					{"invalid": "hit"},
					{"_source": "not_a_map"}
				]
			}
		}`

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := storage.Search(ctx, "test-index", query)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("GetHitsFromResult_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"took": 1,
			"hits": "invalid"
		}`

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := storage.Search(ctx, "test-index", query)
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("ProcessHits_ContextCancelled", func(t *testing.T) {
		storage := setupTestStorage(t)
		es, ok := storage.(*ElasticsearchStorage)
		require.True(t, ok)

		ctx, cancel := context.WithCancel(context.Background())

		hits := []interface{}{
			map[string]interface{}{
				"_source": map[string]interface{}{
					"title": "Test Doc",
				},
			},
		}

		resultChan := make(chan map[string]interface{})

		// Cancel context before starting ProcessHits
		cancel()

		// Process hits with cancelled context
		es.ProcessHits(ctx, hits, resultChan)

		// Try to receive from channel - should not get any results
		select {
		case result, ok := <-resultChan:
			if ok {
				t.Errorf("Received unexpected result after context cancellation: %v", result)
			}
		case <-time.After(100 * time.Millisecond):
			// Success - no results received
		}
	})

	t.Run("HandleScrollResponse_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`

		resultChan := make(chan map[string]interface{})
		searchRes, err := esClient.Search(
			esClient.Search.WithContext(ctx),
			esClient.Search.WithIndex("test-index"),
		)
		require.NoError(t, err)

		scrollID, err := storage.HandleScrollResponse(ctx, searchRes, resultChan)
		assert.Error(t, err)
		assert.Empty(t, scrollID)
	})

	t.Run("HandleScrollResponse_MissingScrollID", func(t *testing.T) {
		mockTransport.Response = `{
			"hits": {
				"hits": [
					{
						"_source": {
							"title": "Test Doc"
						}
					}
				]
			}
		}`

		// Use a buffered channel to prevent blocking
		resultChan := make(chan map[string]interface{}, 1)

		searchRes, err := esClient.Search(
			esClient.Search.WithContext(ctx),
			esClient.Search.WithIndex("test-index"),
		)
		require.NoError(t, err)
		defer searchRes.Body.Close()

		// Create a done channel to signal completion
		done := make(chan struct{})
		go func() {
			defer close(done)
			scrollID, err := storage.HandleScrollResponse(ctx, searchRes, resultChan)
			assert.Error(t, err)
			assert.Empty(t, scrollID)
			close(resultChan) // Close the channel after error
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Success - operation completed
		case <-time.After(100 * time.Millisecond):
			t.Error("Test timed out")
		}
	})

	t.Run("ScrollSearch_NextScrollError", func(t *testing.T) {
		// First response successful
		mockTransport.Response = `{
			"hits": {
				"hits": [
					{
						"_source": {
							"title": "Test Doc"
						}
					}
				]
			},
			"_scroll_id": "test_scroll_id"
		}`

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		// Set up RequestFunc to return error on second request
		var requestCount int
		mockTransport.RequestFunc = func(req *http.Request) (*http.Response, error) {
			requestCount++
			if requestCount > 1 {
				return nil, fmt.Errorf("scroll error")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(mockTransport.Response)),
				Header:     make(http.Header),
			}, nil
		}

		resultChan, err := storage.ScrollSearch(ctx, "test-index", query, 100)
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

		// Read results until channel is closed due to error
		for range resultChan {
			// Should only get one result before error
		}

		// Reset RequestFunc
		mockTransport.RequestFunc = nil
	})

	t.Run("NewStorage_ConnectionError", func(t *testing.T) {
		mockTransport.Error = fmt.Errorf("connection error")

		result, err := NewStorage(&config.Config{
			Elasticsearch: config.ElasticsearchConfig{
				URL: "http://localhost:9200",
			},
		}, log)
		assert.Error(t, err)
		assert.Equal(t, Result{}, result)

		mockTransport.Error = nil
	})

	t.Run("UpdateDocument_InvalidJSON", func(t *testing.T) {
		// Create an update with an unserializable value
		update := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		err := storage.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.Error(t, err)
	})

	t.Run("BulkIndex_InvalidDocument", func(t *testing.T) {
		invalidDoc := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		docs := []interface{}{invalidDoc}
		err := storage.BulkIndex(ctx, "test-index", docs)
		assert.Error(t, err)
	})

	t.Run("IndexDocument_InvalidDocument", func(t *testing.T) {
		doc := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		err := storage.IndexDocument(ctx, "test-index", "test-id", doc)
		assert.Error(t, err)
	})

	t.Run("Search_InvalidQuery", func(t *testing.T) {
		query := map[string]interface{}{
			"query": make(chan int), // Cannot be marshaled to JSON
		}
		results, err := storage.Search(ctx, "test-index", query)
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("CreateIndex_InvalidMapping", func(t *testing.T) {
		mapping := map[string]interface{}{
			"settings": make(chan int), // Cannot be marshaled to JSON
		}
		err := storage.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)
	})

	t.Run("DeleteDocument_ErrorResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "document_missing_exception",
				"reason": "document not found"
			},
			"status": 404
		}`
		mockTransport.StatusCode = http.StatusNotFound

		err := storage.DeleteDocument(ctx, "test-index", "nonexistent-id")
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("Search_ErrorResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "index_not_found_exception",
				"reason": "no such index"
			},
			"status": 404
		}`
		mockTransport.StatusCode = http.StatusNotFound

		results, err := storage.Search(ctx, "nonexistent-index", map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		})
		assert.Error(t, err)
		assert.Nil(t, results)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("ProcessHits_EmptyHits", func(t *testing.T) {
		hits := []interface{}{}
		resultChan := make(chan map[string]interface{}, 1)

		done := make(chan struct{})
		go func() {
			defer close(done)
			storage.ProcessHits(ctx, hits, resultChan)
			close(resultChan)
		}()

		var results []map[string]interface{}
		for result := range resultChan {
			results = append(results, result)
		}
		assert.Empty(t, results)

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for processHits to complete")
		}
	})

	t.Run("NewStorage_InvalidConfig", func(t *testing.T) {
		invalidCfg := &config.Config{
			Elasticsearch: config.ElasticsearchConfig{
				URL: "://invalid-url",
			},
		}
		result, err := NewStorage(invalidCfg, log)
		assert.Error(t, err)
		assert.Equal(t, Result{}, result)
	})

	t.Run("TestConnection_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		err := storage.TestConnection(ctx)
		assert.Error(t, err)
	})

	t.Run("ScrollSearch_InvalidQuery", func(t *testing.T) {
		invalidQuery := map[string]interface{}{
			"invalid": make(chan int), // Cannot be marshaled to JSON
		}
		resultChan, err := storage.ScrollSearch(ctx, "test-index", invalidQuery, 100)
		assert.Error(t, err)
		assert.Nil(t, resultChan)
	})

	t.Run("UpdateDocument_ErrorResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "version_conflict_engine_exception",
				"reason": "version conflict"
			},
			"status": 409
		}`
		mockTransport.StatusCode = http.StatusConflict

		update := map[string]interface{}{
			"title": "Updated Title",
		}
		err := storage.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("BulkIndex_EmptyDocuments", func(t *testing.T) {
		err := storage.BulkIndex(ctx, "test-index", []interface{}{})
		assert.NoError(t, err)
	})

	t.Run("CreateIndex_ErrorResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "resource_already_exists_exception",
				"reason": "index already exists"
			},
			"status": 400
		}`
		mockTransport.StatusCode = http.StatusBadRequest

		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
				},
			},
		}
		err := storage.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("DeleteIndex_NonexistentIndex", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "index_not_found_exception",
				"reason": "no such index"
			},
			"status": 404
		}`
		mockTransport.StatusCode = http.StatusNotFound

		err := storage.DeleteIndex(ctx, "nonexistent-index")
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("Search_IsErrorResponse", func(t *testing.T) {
		mockTransport.Response = `{
			"error": {
				"type": "search_exception",
				"reason": "search error"
			},
			"status": 500
		}`
		mockTransport.StatusCode = http.StatusInternalServerError

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := storage.Search(ctx, "test-index", query)
		assert.Error(t, err)
		assert.Nil(t, results)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("DeleteDocument_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError
		err := storage.DeleteDocument(ctx, "test-index", "test-id")
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("UpdateDocument_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError
		update := map[string]interface{}{
			"title": "Updated Title",
		}
		err := storage.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("BulkIndex_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError

		docs := []interface{}{
			map[string]interface{}{"title": "Doc 1"},
		}
		err := storage.BulkIndex(ctx, "test-index", docs)
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("ScrollSearch_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError
		mockTransport.RequestFunc = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
				Header:     make(http.Header),
			}, nil
		}

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resultChan, err := storage.ScrollSearch(ctx, "test-index", query, 100)
		assert.Error(t, err)
		assert.Nil(t, resultChan)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
		mockTransport.RequestFunc = nil
	})

	t.Run("ProcessHits_InvalidSource", func(t *testing.T) {
		hits := []interface{}{
			map[string]interface{}{
				"_source": "invalid source", // Not a map
			},
		}
		resultChan := make(chan map[string]interface{}, 1)
		done := make(chan struct{})

		go func() {
			defer close(done)
			storage.ProcessHits(ctx, hits, resultChan)
			close(resultChan)
		}()

		var results []map[string]interface{}
		for result := range resultChan {
			results = append(results, result)
		}
		assert.Empty(t, results)

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for processHits to complete")
		}
	})

	t.Run("NewStorage_InvalidTransport", func(t *testing.T) {
		invalidTransport := &http.Transport{}

		result, err := NewStorage(&config.Config{
			Elasticsearch: config.ElasticsearchConfig{
				URL: "http://localhost:9200",
			},
			Crawler: config.CrawlerConfig{
				Transport: invalidTransport,
			},
		}, log)
		assert.Error(t, err)
		assert.Equal(t, Result{}, result)
	})

	t.Run("CreateIndex_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError

		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
				},
			},
		}
		err := storage.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})

	t.Run("DeleteIndex_InvalidResponse", func(t *testing.T) {
		mockTransport.Response = `invalid json`
		mockTransport.StatusCode = http.StatusInternalServerError

		err := storage.DeleteIndex(ctx, "test-index")
		assert.Error(t, err)

		// Reset transport
		mockTransport.StatusCode = http.StatusOK
		mockTransport.Response = successResponse
	})
}

// MockElasticsearchClient is a mock of the Elasticsearch client
type MockElasticsearchClient struct {
	mock.Mock
}

func (m *MockElasticsearchClient) Search(ctx context.Context, index string, query interface{}) (*esapi.Response, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func (m *MockElasticsearchClient) Index(ctx context.Context, index string, body interface{}) (*esapi.Response, error) {
	args := m.Called(ctx, index, body)
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func (m *MockElasticsearchClient) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Add other methods as needed

func TestNewStorageWithMock(t *testing.T) {
	// Create a mock Elasticsearch client
	mockClient := new(MockElasticsearchClient)

	// Set up expectations on the mock client
	mockClient.On("Search", mock.Anything, "test-index", mock.Anything).Return(&esapi.Response{}, nil)
	mockClient.On("Index", mock.Anything, "test-index", mock.Anything).Return(&esapi.Response{}, nil)
	mockClient.On("TestConnection", mock.Anything).Return(nil)

	// Create test logger using the new mock logger
	log := logger.NewMockCustomLogger()

	// Create storage instance
	storage, err := NewStorage(&config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
		Crawler: config.CrawlerConfig{
			Transport: http.DefaultTransport,
		},
	}, log)
	require.NoError(t, err)
	require.NotNil(t, storage.Storage)
}
