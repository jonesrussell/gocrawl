package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

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
		docs := []interface{}{
			map[string]interface{}{"title": "Doc 1"},
			map[string]interface{}{"title": "Doc 2"},
		}
		err := es.BulkIndex(ctx, "test-index", docs)
		assert.NoError(t, err)
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
}
