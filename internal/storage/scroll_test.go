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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrollSearch(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	// First response with scroll ID and hits
	firstResponse := `{
		"took": 1,
		"_scroll_id": "test_scroll_id",
		"hits": {
			"total": {"value": 3, "relation": "eq"},
			"hits": [
				{
					"_source": {"field1": "value1"},
					"_id": "1"
				},
				{
					"_source": {"field1": "value2"},
					"_id": "2"
				},
				{
					"_source": {"field1": "value3"},
					"_id": "3"
				}
			]
		}
	}`

	// Second response indicating end of scroll with error
	endResponse := `{
		"error": {
			"type": "search_phase_execution_exception",
			"reason": "no search context found"
		},
		"status": 404
	}`

	var requestCount int
	mockTransport := &mockTransport{
		Response:   firstResponse,
		StatusCode: http.StatusOK,
		RequestFunc: func(req *http.Request) (*http.Response, error) {
			requestCount++
			header := http.Header{}
			header.Add("X-Elastic-Product", "Elasticsearch")
			header.Add("Content-Type", "application/json")

			var responseBody string
			statusCode := http.StatusOK

			if requestCount == 1 {
				responseBody = firstResponse
			} else {
				responseBody = endResponse
				statusCode = http.StatusNotFound
			}

			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header:     header,
			}, nil
		},
	}

	// Update storage with mock transport
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
		Addresses: []string{storage.opts.URL},
	})
	require.NoError(t, err)
	storage.ESClient = client

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	resultChan, err := storage.ScrollSearch(ctx, "test_scroll", query, 1)
	require.NoError(t, err)

	var results []map[string]interface{}
	done := make(chan struct{})

	go func() {
		defer close(done)
		for doc := range resultChan {
			results = append(results, doc)
		}
	}()

	select {
	case <-done:
		assert.Len(t, results, 3, "Expected 3 documents from scroll")
		for i, result := range results {
			expectedValue := fmt.Sprintf("value%d", i+1)
			assert.Equal(t, expectedValue, result["field1"], "Unexpected value for document %d", i+1)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for scroll results")
	}
}

func TestScrollSearchWithCancel(t *testing.T) {
	storage := setupTestStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	responseBody := `{
		"_scroll_id": "test_scroll_id",
		"hits": {
			"hits": [
				{
					"_source": {"field1": "value1"},
					"_id": "1"
				}
			]
		}
	}`
	mockTransport := &mockTransport{
		Response:   responseBody,
		StatusCode: http.StatusOK,
		RequestFunc: func(req *http.Request) (*http.Response, error) {
			header := http.Header{}
			header.Add("X-Elastic-Product", "Elasticsearch")
			header.Add("Content-Type", "application/json")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header:     header,
			}, nil
		},
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
		Addresses: []string{storage.opts.URL},
	})
	require.NoError(t, err)
	storage.ESClient = client

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	resultChan, err := storage.ScrollSearch(ctx, "test_scroll_cancel", query, 1)
	require.NoError(t, err)

	var results []map[string]interface{}
	for doc := range resultChan {
		results = append(results, doc)
		cancel()
		break
	}

	assert.Equal(t, 1, len(results))
}

func TestProcessHits(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	hits := []interface{}{
		map[string]interface{}{
			"_source": map[string]interface{}{
				"title": "Test Doc",
			},
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
	assert.Len(t, results, 1)
}

func TestHandleScrollResponse(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	// Create test response with proper format
	responseBody := `{
		"took": 1,
		"_scroll_id": "test_scroll_id",
		"hits": {
			"total": {"value": 1, "relation": "eq"},
			"hits": [
				{
					"_source": {"field1": "value1"},
					"_id": "1"
				}
			]
		}
	}`

	mockTransport := &mockTransport{
		Response:   responseBody,
		StatusCode: http.StatusOK,
		RequestFunc: func(req *http.Request) (*http.Response, error) {
			header := http.Header{}
			header.Add("X-Elastic-Product", "Elasticsearch")
			header.Add("Content-Type", "application/json")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header:     header,
			}, nil
		},
	}

	// Create client with mock transport
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
		Addresses: []string{storage.opts.URL},
	})
	require.NoError(t, err)
	storage.ESClient = client

	// Create response for testing
	searchRes, err := storage.ESClient.Search(
		storage.ESClient.Search.WithContext(ctx),
		storage.ESClient.Search.WithIndex("test_index"),
	)
	require.NoError(t, err)

	resultChan := make(chan map[string]interface{}, 1)
	scrollID, err := storage.HandleScrollResponse(ctx, searchRes, resultChan)
	require.NoError(t, err)
	assert.Equal(t, "test_scroll_id", scrollID)

	// Check that we got the expected result
	result := <-resultChan
	assert.Equal(t, "value1", result["field1"])
}
