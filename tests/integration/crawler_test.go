// Package integration_test provides integration tests for GoCrawl.
// These tests verify component interactions and end-to-end workflows.
package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/tests/helpers"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ElasticsearchStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Elasticsearch container
	esContainer, err := helpers.StartElasticsearch(ctx)
	require.NoError(t, err, "failed to start Elasticsearch container")
	defer func() {
		_ = esContainer.Stop(ctx)
	}()

	// Create test config
	testConfig := &config.Config{
		Elasticsearch: &elasticsearch.Config{
			Addresses: esContainer.GetAddresses(),
			Username:  "elastic",
			Password:  "changeme",
		},
	}

	// Create logger
	testLogger := logger.NewNoOpLogger()

	// Create Elasticsearch client
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: testConfig,
		Logger: testLogger,
	})
	require.NoError(t, err, "failed to create Elasticsearch client")

	// Create storage
	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: testConfig,
		Logger: testLogger,
		Client: clientResult.Client,
	})
	require.NoError(t, err, "failed to create storage")

	storageClient := storageResult.Storage

	// Create test index
	indexName := "test_integration_index"
	err = storageClient.CreateIndex(ctx, indexName, nil)
	require.NoError(t, err, "failed to create index")

	// Wait for index to be ready
	helpers.WaitForIndexReady(t, storageClient, ctx, indexName, 10*time.Second)

	// Index a test document
	testDoc := map[string]any{
		"title":   "Test Document",
		"content": "This is a test document for integration testing",
	}
	err = storageClient.IndexDocument(ctx, indexName, "test-doc-1", testDoc)
	require.NoError(t, err, "failed to index document")

	// Wait a bit for indexing to complete
	time.Sleep(1 * time.Second)

	// Verify document was indexed
	helpers.AssertDocumentIndexed(t, storageClient, ctx, indexName, "test-doc-1")

	// Verify document count
	helpers.AssertDocumentCount(t, storageClient, ctx, indexName, 1)
}
