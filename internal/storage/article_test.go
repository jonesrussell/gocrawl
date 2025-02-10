package storage

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateArticlesIndex(t *testing.T) {
	es := setupTestStorage(t)
	ctx := context.Background()

	// Mock successful index creation
	mockTransport := &mockTransport{
		Response:   `{"acknowledged": true}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	err := es.CreateArticlesIndex(ctx)
	require.NoError(t, err)
}

func TestIndexArticle(t *testing.T) {
	es := setupTestStorage(t)
	ctx := context.Background()

	// Mock successful index response
	mockTransport := &mockTransport{
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

	err := es.IndexArticle(ctx, article)
	require.NoError(t, err)
}

func TestBulkIndexArticles(t *testing.T) {
	es := setupTestStorage(t)
	ctx := context.Background()

	// Mock successful bulk index response
	mockTransport := &mockTransport{
		Response: `{
			"took": 30,
			"errors": false,
			"items": [
				{
					"index": {
						"_index": "articles",
						"_id": "test-1",
						"_version": 1,
						"result": "created",
						"status": 201
					}
				},
				{
					"index": {
						"_index": "articles",
						"_id": "test-2",
						"_version": 1,
						"result": "created",
						"status": 201
					}
				}
			]
		}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	articles := []*models.Article{
		{
			ID:            "test-1",
			Title:         "Test Article 1",
			Body:          "Test content 1",
			Source:        "https://example.com/1",
			PublishedDate: time.Now(),
			Tags:          []string{"test", "article"},
		},
		{
			ID:            "test-2",
			Title:         "Test Article 2",
			Body:          "Test content 2",
			Source:        "https://example.com/2",
			PublishedDate: time.Now(),
			Tags:          []string{"test", "article"},
		},
	}

	err := es.BulkIndexArticles(ctx, articles)
	require.NoError(t, err)
}

func TestSearchArticles(t *testing.T) {
	es := setupTestStorage(t)
	ctx := context.Background()

	// Mock successful search response
	mockTransport := &mockTransport{
		Response: `{
			"took": 1,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{
						"_source": {
							"id": "test-1",
							"title": "Golang Testing",
							"body": "How to write tests in Go",
							"source": "https://example.com/1",
							"published_date": "2024-03-20T12:00:00Z",
							"tags": ["golang", "testing"]
						}
					},
					{
						"_source": {
							"id": "test-2",
							"title": "Elasticsearch Guide",
							"body": "Using Elasticsearch with Go",
							"source": "https://example.com/2",
							"published_date": "2024-03-20T12:00:00Z",
							"tags": ["elasticsearch", "golang"]
						}
					}
				]
			}
		}`,
		StatusCode: http.StatusOK,
	}
	es.ESClient.Transport = mockTransport

	t.Run("search by title", func(t *testing.T) {
		results, err := es.SearchArticles(ctx, "Golang", 10)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Golang Testing", results[0].Title)
	})

	t.Run("search by content", func(t *testing.T) {
		results, err := es.SearchArticles(ctx, "Elasticsearch", 10)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Elasticsearch Guide", results[1].Title)
	})

	t.Run("search by tag", func(t *testing.T) {
		results, err := es.SearchArticles(ctx, "golang", 10)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})
}
