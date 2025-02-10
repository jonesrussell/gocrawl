package storage

import (
	"net/http"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
)

// setupTestStorage creates a test storage instance with mocked transport
// This is used by multiple test files in the package
func setupTestStorage(t *testing.T) *ElasticsearchStorage {
	// Create mock transport
	mockTransport := &mockTransport{
		Response: `{
			"name": "test-node",
			"cluster_name": "test-cluster",
			"version": {
				"number": "8.0.0"
			}
		}`,
		StatusCode: http.StatusOK,
	}

	// Create a test config
	cfg := &config.Config{
		ElasticURL: "http://localhost:9200",
	}

	// Create elasticsearch client with mock transport
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
		Addresses: []string{cfg.ElasticURL},
	})
	require.NoError(t, err)

	// Create a test logger
	log := logger.NewMockCustomLogger()

	// Create storage instance with mocked client
	storage := &ElasticsearchStorage{
		ESClient: esClient,
		Logger:   log,
		opts:     DefaultOptions(),
	}

	return storage
}
