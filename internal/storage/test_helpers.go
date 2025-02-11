package storage

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
)

// testConfig returns a test configuration
func testConfig() *config.Config {
	return &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
		App: config.AppConfig{
			LogLevel: "info",
		},
	}
}

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

// CreateTestStorage creates a new storage instance for testing with full configuration
func CreateTestStorage(t *testing.T) Storage {
	t.Helper()

	// Create test logger
	log := logger.NewMockCustomLogger()

	// Create test config
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL:      getTestElasticURL(),
			Password: "test-password",
			APIKey:   "test-api-key",
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
