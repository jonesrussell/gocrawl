package storage

import (
	"context"
	"net/http"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStorage creates a new storage instance for testing
func setupTestStorage(t *testing.T) Interface {
	t.Helper()

	// Create storage instance
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"}, // Ensure this matches your Elasticsearch URL
	})
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %v", err)
	}

	storageInstance, err := NewStorage(esClient)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	return storageInstance
}

func TestScrollSearch(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Use default batch size for tests
	const batchSize = 100

	// Mock successful scroll response
	mockTransport := &MockTransport{
		Response: `{
			"_scroll_id": "test_scroll_id",
			"took": 1,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{
						"_source": {
							"title": "Test Document 1"
						}
					},
					{
						"_source": {
							"title": "Test Document 2"
						}
					}
				]
			}
		}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	// Test scroll search
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	resultChan, err := es.ScrollSearch(ctx, "test-index", query, batchSize)
	require.NoError(t, err)
	require.NotNil(t, resultChan)

	// Verify results
	var results []map[string]interface{}
	for result := range resultChan {
		results = append(results, result)
	}

	assert.Len(t, results, 2)
	assert.Equal(t, "Test Document 1", results[0]["title"])
	assert.Equal(t, "Test Document 2", results[1]["title"])
}

func TestScrollSearch_Error(t *testing.T) {
	ctx := context.Background()
	storage := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storage.(*ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Use default batch size for tests
	const batchSize = 100

	// Mock error response
	mockTransport := &MockTransport{
		Response:   `{"error": "test error"}`,
		StatusCode: http.StatusInternalServerError,
	}
	es.ESClient.Transport = mockTransport

	// Test scroll search with error
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	resultChan, err := es.ScrollSearch(ctx, "test-index", query, batchSize)
	assert.Error(t, err)
	assert.Nil(t, resultChan)
}

func TestProcessHits(t *testing.T) {
	ctx := context.Background()
	storage := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storage.(*ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	hits := []interface{}{
		map[string]interface{}{
			"_source": map[string]interface{}{
				"title": "Test Document",
			},
		},
	}

	resultChan := make(chan map[string]interface{}, 1)
	go func() {
		es.ProcessHits(ctx, hits, resultChan)
		close(resultChan)
	}()

	var results []map[string]interface{}
	for result := range resultChan {
		results = append(results, result)
	}

	assert.Len(t, results, 1)
	assert.Equal(t, "Test Document", results[0]["title"])
}

func TestHandleScrollResponse(t *testing.T) {
	storage := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storage.(*ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Mock scroll response
	mockTransport := &MockTransport{
		Response: `{
			"_scroll_id": "test_scroll_id",
			"hits": {
				"hits": [
					{
						"_source": {
							"title": "Test Document"
						}
					}
				]
			}
		}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	// Create test response
	resp := mockTransport.createESResponse()

	// Create result channel
	resultChan := make(chan map[string]interface{}, 1)
	defer close(resultChan)

	// Test response handling
	scrollID, err := es.HandleScrollResponse(context.Background(), resp, resultChan)
	require.NoError(t, err)
	assert.Equal(t, "test_scroll_id", scrollID)

	// Verify a result was sent to the channel
	result := <-resultChan
	assert.NotNil(t, result)
	assert.Equal(t, "Test Document", result["title"])
}
