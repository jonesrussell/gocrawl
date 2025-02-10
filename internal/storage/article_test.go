package storage

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateArticlesIndex(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	// Test successful index creation
	err := storage.CreateArticlesIndex(ctx)
	require.NoError(t, err)

	// Verify index exists
	res, err := storage.ESClient.Indices.Exists([]string{"articles"})
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestIndexArticle(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	// Create test article
	article := &models.Article{
		ID:            "test-1",
		Title:         "Test Article",
		Body:          "Test content",
		Source:        "https://example.com",
		PublishedDate: time.Now(),
	}

	// Test indexing
	err := storage.IndexArticle(ctx, article)
	require.NoError(t, err)

	// Verify article exists
	res, err := storage.ESClient.Get("articles", article.ID)
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestBulkIndexArticles(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	// Create test articles
	articles := []*models.Article{
		{
			ID:            "bulk-1",
			Title:         "Bulk Article 1",
			Body:          "Content 1",
			Source:        "https://example.com",
			PublishedDate: time.Now(),
		},
		{
			ID:            "bulk-2",
			Title:         "Bulk Article 2",
			Body:          "Content 2",
			Source:        "https://example.com",
			PublishedDate: time.Now(),
		},
	}

	// Test bulk indexing
	err := storage.BulkIndexArticles(ctx, articles)
	require.NoError(t, err)

	// Verify articles exist
	for _, article := range articles {
		res, err := storage.ESClient.Get("articles", article.ID)
		require.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
	}
}
