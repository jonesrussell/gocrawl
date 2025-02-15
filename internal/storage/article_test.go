package storage_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTransport is a mock implementation of http.RoundTripper
type MockTransport struct {
	Response   string
	StatusCode int
}

// RoundTrip implements the http.RoundTripper interface
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body:       io.NopCloser(strings.NewReader(m.Response)),
		StatusCode: m.StatusCode,
	}, nil
}

// Perform implements elastictransport.Interface
func (m *MockTransport) Perform(req *http.Request) (*http.Response, error) {
	return m.RoundTrip(req)
}

// setupTestStorage creates a new storage instance for testing
func setupTestStorage(t *testing.T) storage.Interface {
	t.Helper()

	// Create Elasticsearch client for testing
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"}, // Ensure this matches your Elasticsearch URL
	})
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %v", err)
	}

	// Create storageInstance instance
	storageInstance, err := storage.NewStorage(esClient)
	require.NoError(t, err, "Failed to create test storage")
	require.NotNil(t, storageInstance, "Storage instance should not be nil")

	return storageInstance
}

func TestCreateArticlesIndex(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Mock successful index creation
	mockTransport := &MockTransport{
		Response:   `{"acknowledged": true}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	// Create index mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	err := es.CreateIndex(ctx, "articles", mapping)
	require.NoError(t, err)
}

func TestIndexArticle(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Mock successful index response
	mockTransport := &MockTransport{
		Response: `{
			"_index": "articles",
			"_id": "test-1",
			"_version": 1,
			"result": "created",
			"_shards": {"total": 1, "successful": 1, "failed": 0}
		}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	article := &models.Article{
		ID:            "test-1",
		Title:         "Test Article",
		Body:          "Test content",
		Source:        "https://example.com",
		PublishedDate: time.Now(),
	}

	err := es.IndexDocument(ctx, "articles", article.ID, article)
	require.NoError(t, err)
}

func TestBulkIndexArticles(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Test successful bulk indexing
	articles := []*models.Article{
		{
			ID:     "1",
			Title:  "Test Article 1",
			Body:   "Test Body 1",
			Source: "http://test.com/1",
		},
		{
			ID:     "2",
			Title:  "Test Article 2",
			Body:   "Test Body 2",
			Source: "http://test.com/2",
		},
	}

	err := es.BulkIndexArticles(ctx, articles)
	assert.NoError(t, err)
}

func TestSearchArticles(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Test successful search
	articles, err := es.SearchArticles(ctx, "test", 10)
	assert.NoError(t, err)
	assert.NotNil(t, articles)
}

func TestBulkIndexArticles_EmptyList(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Test with empty list
	err := es.BulkIndexArticles(ctx, []*models.Article{})
	assert.NoError(t, err)
}

func TestSearchArticles_NoResults(t *testing.T) {
	ctx := context.Background()
	storageInstance := setupTestStorage(t)

	// Type assertion to get the concrete type
	es, ok := storageInstance.(*storage.ElasticsearchStorage)
	require.True(t, ok, "Storage should be of type *ElasticsearchStorage")

	// Test with no results
	articles, err := es.SearchArticles(ctx, "nonexistent", 10)
	assert.NoError(t, err)
	assert.Empty(t, articles)
}
