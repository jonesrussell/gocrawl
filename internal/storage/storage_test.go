package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransport struct {
	Response    string
	StatusCode  int
	Error       error
	RequestFunc func(*http.Request) (*http.Response, error)
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Error != nil {
		return nil, t.Error
	}

	if t.RequestFunc != nil {
		return t.RequestFunc(req)
	}

	// Add required Elasticsearch response headers
	header := make(http.Header)
	header.Add("X-Elastic-Product", "Elasticsearch") // This is crucial
	header.Add("Content-Type", "application/json")

	response := &http.Response{
		StatusCode: t.StatusCode,
		Body:       io.NopCloser(strings.NewReader(t.Response)),
		Header:     header,
	}
	return response, nil
}

func TestNewStorage(t *testing.T) {
	log := logger.NewMockCustomLogger()

	// Mock successful response
	successResponse := `{
		"name" : "test_node",
		"cluster_name" : "test_cluster",
		"version" : {
			"number" : "8.0.0"
		}
	}`

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				ElasticURL: "http://localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "missing elastic URL",
			cfg: &config.Config{
				ElasticURL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create elasticsearch client with mock transport
			transport := &mockTransport{
				Response:   successResponse,
				StatusCode: http.StatusOK,
			}

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Transport: transport,
				Addresses: []string{tt.cfg.ElasticURL},
			})
			require.NoError(t, err)

			// Create storage with mocked client
			result, err := NewStorageWithClient(tt.cfg, log, esClient)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result.Storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result.Storage)
			}
		})
	}
}

func TestElasticsearchStorage_Operations(t *testing.T) {
	ctx := context.Background()
	log := logger.NewMockCustomLogger()

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

	transport := &mockTransport{
		Response:   successResponse,
		StatusCode: http.StatusOK,
	}

	cfg := &config.Config{
		ElasticURL: "http://localhost:9200",
		IndexName:  "test-index",
	}

	// Create elasticsearch client with mock transport
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: transport,
		Addresses: []string{cfg.ElasticURL},
	})
	require.NoError(t, err)

	// Create storage with mocked client
	storage, err := NewStorageWithClient(cfg, log, esClient)
	require.NoError(t, err)
	require.NotNil(t, storage.Storage)

	es := storage.Storage.(*ElasticsearchStorage)

	t.Run("IndexDocument", func(t *testing.T) {
		// Update transport response for index operation
		transport.Response = `{
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
		err := es.IndexDocument(ctx, "test-index", "test-id", doc)
		assert.NoError(t, err)
	})

	t.Run("Search", func(t *testing.T) {
		// Update transport response for search operation
		transport.Response = successResponse

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := es.Search(ctx, "test-index", query)
		assert.NoError(t, err)
		assert.NotNil(t, results)
	})

	t.Run("BulkIndex", func(t *testing.T) {
		// Update transport response for bulk operation
		transport.Response = `{
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
		err := es.BulkIndex(ctx, "test-index", docs)
		assert.NoError(t, err)

		// Test error case
		transport.Error = fmt.Errorf("bulk index error")
		err = es.BulkIndex(ctx, "test-index", docs)
		assert.Error(t, err)
		transport.Error = nil
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		update := map[string]interface{}{
			"title": "Updated Title",
		}
		err := es.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.NoError(t, err)
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := es.DeleteDocument(ctx, "test-index", "test-id")
		assert.NoError(t, err)
	})

	t.Run("ScrollSearch", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resultChan, err := es.ScrollSearch(ctx, "test-index", query, 100)
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

		// Read some results
		for result := range resultChan {
			assert.NotNil(t, result)
		}
	})

	t.Run("TestConnection", func(t *testing.T) {
		transport.Response = `{
			"name" : "node-1",
			"cluster_name" : "elasticsearch",
			"version" : {
				"number" : "8.0.0"
			}
		}`
		err := es.TestConnection(ctx)
		assert.NoError(t, err)

		// Test error case
		transport.Error = fmt.Errorf("connection error")
		err = es.TestConnection(ctx)
		assert.Error(t, err)
		transport.Error = nil
	})

	t.Run("CreateIndex", func(t *testing.T) {
		transport.Response = `{
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
		err := es.CreateIndex(ctx, "test-index", mapping)
		assert.NoError(t, err)

		// Test error case
		transport.Error = fmt.Errorf("create index error")
		err = es.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)
		transport.Error = nil
	})

	t.Run("DeleteIndex", func(t *testing.T) {
		transport.Response = `{
			"acknowledged": true
		}`
		err := es.DeleteIndex(ctx, "test-index")
		assert.NoError(t, err)

		// Test error case
		transport.Error = fmt.Errorf("delete index error")
		err = es.DeleteIndex(ctx, "test-index")
		assert.Error(t, err)
		transport.Error = nil
	})

	t.Run("ScrollSearch_Error", func(t *testing.T) {
		// Set error response for initial search
		transport.Response = `{
			"error": {
				"type": "search_phase_execution_exception",
				"reason": "scroll error"
			},
			"status": 500
		}`
		transport.StatusCode = http.StatusInternalServerError

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resultChan, err := es.ScrollSearch(ctx, "test-index", query, 100)
		assert.Error(t, err)
		assert.Nil(t, resultChan)

		// Reset transport for subsequent tests
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})

	t.Run("ProcessHits_InvalidHit", func(t *testing.T) {
		transport.Response = `{
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
		results, err := es.Search(ctx, "test-index", query)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("GetHitsFromResult_InvalidResponse", func(t *testing.T) {
		transport.Response = `{
			"took": 1,
			"hits": "invalid"
		}`

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		results, err := es.Search(ctx, "test-index", query)
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("ProcessHits_ContextCancelled", func(t *testing.T) {
		// Create a cancelled context
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		hits := []interface{}{
			map[string]interface{}{
				"_source": map[string]interface{}{
					"title": "Test Doc",
				},
			},
		}

		resultChan := make(chan map[string]interface{})
		done := make(chan struct{})

		go func() {
			defer close(done)
			es.processHits(cancelCtx, hits, resultChan)
			close(resultChan)
		}()

		// Use a timeout to avoid hanging if the test fails
		select {
		case result, ok := <-resultChan:
			if ok {
				t.Errorf("Received unexpected result: %v", result)
			}
		case <-done:
			// Success - channel was closed without sending results
		case <-time.After(time.Second):
			t.Error("Test timed out")
		}
	})

	t.Run("HandleScrollResponse_InvalidResponse", func(t *testing.T) {
		transport.Response = `invalid json`

		resultChan := make(chan map[string]interface{})
		searchRes, err := es.ESClient.Search(
			es.ESClient.Search.WithContext(ctx),
			es.ESClient.Search.WithIndex("test-index"),
		)
		require.NoError(t, err)

		scrollID, err := es.handleScrollResponse(ctx, searchRes, resultChan)
		assert.Error(t, err)
		assert.Empty(t, scrollID)
	})

	t.Run("HandleScrollResponse_MissingScrollID", func(t *testing.T) {
		transport.Response = `{
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

		searchRes, err := es.ESClient.Search(
			es.ESClient.Search.WithContext(ctx),
			es.ESClient.Search.WithIndex("test-index"),
		)
		require.NoError(t, err)
		defer searchRes.Body.Close()

		// Create a done channel to signal completion
		done := make(chan struct{})
		go func() {
			defer close(done)
			scrollID, err := es.handleScrollResponse(ctx, searchRes, resultChan)
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
		transport.Response = `{
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
		transport.RequestFunc = func(req *http.Request) (*http.Response, error) {
			requestCount++
			if requestCount > 1 {
				return nil, fmt.Errorf("scroll error")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(transport.Response)),
				Header:     make(http.Header),
			}, nil
		}

		resultChan, err := es.ScrollSearch(ctx, "test-index", query, 100)
		assert.NoError(t, err)
		assert.NotNil(t, resultChan)

		// Read results until channel is closed due to error
		for range resultChan {
			// Should only get one result before error
		}

		// Reset RequestFunc
		transport.RequestFunc = nil
	})

	t.Run("NewStorage_ConnectionError", func(t *testing.T) {
		transport.Error = fmt.Errorf("connection error")

		result, err := NewStorage(&config.Config{
			ElasticURL: "http://localhost:9200",
		}, log)
		assert.Error(t, err)
		assert.Equal(t, Result{}, result)

		transport.Error = nil
	})

	t.Run("UpdateDocument_InvalidJSON", func(t *testing.T) {
		// Create an update with an unserializable value
		update := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		err := es.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.Error(t, err)
	})

	t.Run("BulkIndex_InvalidDocument", func(t *testing.T) {
		invalidDoc := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		docs := []interface{}{invalidDoc}
		err := es.BulkIndex(ctx, "test-index", docs)
		assert.Error(t, err)
	})

	t.Run("IndexDocument_InvalidDocument", func(t *testing.T) {
		doc := map[string]interface{}{
			"value": make(chan int), // Cannot be marshaled to JSON
		}
		err := es.IndexDocument(ctx, "test-index", "test-id", doc)
		assert.Error(t, err)
	})

	t.Run("Search_InvalidQuery", func(t *testing.T) {
		query := map[string]interface{}{
			"query": make(chan int), // Cannot be marshaled to JSON
		}
		results, err := es.Search(ctx, "test-index", query)
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("CreateIndex_InvalidMapping", func(t *testing.T) {
		mapping := map[string]interface{}{
			"settings": make(chan int), // Cannot be marshaled to JSON
		}
		err := es.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)
	})

	t.Run("DeleteDocument_ErrorResponse", func(t *testing.T) {
		transport.Response = `{
			"error": {
				"type": "document_missing_exception",
				"reason": "document not found"
			},
			"status": 404
		}`
		transport.StatusCode = http.StatusNotFound

		err := es.DeleteDocument(ctx, "test-index", "nonexistent-id")
		assert.Error(t, err)

		// Reset transport
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})

	t.Run("Search_ErrorResponse", func(t *testing.T) {
		transport.Response = `{
			"error": {
				"type": "index_not_found_exception",
				"reason": "no such index"
			},
			"status": 404
		}`
		transport.StatusCode = http.StatusNotFound

		results, err := es.Search(ctx, "nonexistent-index", map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		})
		assert.Error(t, err)
		assert.Nil(t, results)

		// Reset transport
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})

	t.Run("ProcessHits_EmptyHits", func(t *testing.T) {
		hits := []interface{}{}
		resultChan := make(chan map[string]interface{}, 1)

		done := make(chan struct{})
		go func() {
			defer close(done)
			es.processHits(ctx, hits, resultChan)
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
			ElasticURL: "://invalid-url",
		}
		result, err := NewStorage(invalidCfg, log)
		assert.Error(t, err)
		assert.Equal(t, Result{}, result)
	})

	t.Run("TestConnection_InvalidResponse", func(t *testing.T) {
		transport.Response = `invalid json`
		err := es.TestConnection(ctx)
		assert.Error(t, err)
	})

	t.Run("ScrollSearch_InvalidQuery", func(t *testing.T) {
		invalidQuery := map[string]interface{}{
			"invalid": make(chan int), // Cannot be marshaled to JSON
		}
		resultChan, err := es.ScrollSearch(ctx, "test-index", invalidQuery, 100)
		assert.Error(t, err)
		assert.Nil(t, resultChan)
	})

	t.Run("UpdateDocument_ErrorResponse", func(t *testing.T) {
		transport.Response = `{
			"error": {
				"type": "version_conflict_engine_exception",
				"reason": "version conflict"
			},
			"status": 409
		}`
		transport.StatusCode = http.StatusConflict

		update := map[string]interface{}{
			"title": "Updated Title",
		}
		err := es.UpdateDocument(ctx, "test-index", "test-id", update)
		assert.Error(t, err)

		// Reset transport
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})

	t.Run("BulkIndex_EmptyDocuments", func(t *testing.T) {
		err := es.BulkIndex(ctx, "test-index", []interface{}{})
		assert.NoError(t, err)
	})

	t.Run("CreateIndex_ErrorResponse", func(t *testing.T) {
		transport.Response = `{
			"error": {
				"type": "resource_already_exists_exception",
				"reason": "index already exists"
			},
			"status": 400
		}`
		transport.StatusCode = http.StatusBadRequest

		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
				},
			},
		}
		err := es.CreateIndex(ctx, "test-index", mapping)
		assert.Error(t, err)

		// Reset transport
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})

	t.Run("DeleteIndex_NonexistentIndex", func(t *testing.T) {
		transport.Response = `{
			"error": {
				"type": "index_not_found_exception",
				"reason": "no such index"
			},
			"status": 404
		}`
		transport.StatusCode = http.StatusNotFound

		err := es.DeleteIndex(ctx, "nonexistent-index")
		assert.Error(t, err)

		// Reset transport
		transport.StatusCode = http.StatusOK
		transport.Response = successResponse
	})
}

func TestNewStorage_Errors(t *testing.T) {
	log := logger.NewMockCustomLogger()

	// Create a config with empty URL to trigger error
	cfg := &config.Config{
		ElasticURL: "", // This should trigger an error
	}

	// Create elasticsearch client with mock transport
	transport := &mockTransport{
		Response:   "{}",
		StatusCode: http.StatusOK,
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: transport,
	})
	require.NoError(t, err)

	// Test storage creation with empty URL
	result, err := NewStorageWithClient(cfg, log, esClient)
	assert.Error(t, err)
	assert.Equal(t, Result{}, result) // Should be an empty Result
}
