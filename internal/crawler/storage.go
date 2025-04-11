package crawler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Storage implements the ArticleStorage interface using the underlying storage implementation.
type Storage struct {
	logger    logger.Interface
	storage   storagetypes.Interface
	indexName string
}

// NewStorage creates a new Storage instance.
func NewStorage(
	logger logger.Interface,
	storage storagetypes.Interface,
	indexName string,
) *Storage {
	return &Storage{
		logger:    logger,
		storage:   storage,
		indexName: indexName,
	}
}

// SaveArticle saves an article to storage.
func (s *Storage) SaveArticle(ctx context.Context, article *models.Article) error {
	if article == nil {
		return fmt.Errorf("article is nil")
	}

	if err := s.storage.IndexDocument(ctx, s.indexName, article.ID, article); err != nil {
		s.logger.Error("Failed to save article",
			"error", err,
			"articleID", article.ID,
			"url", article.Source)
		return fmt.Errorf("failed to save article: %w", err)
	}

	s.logger.Debug("Saved article",
		"articleID", article.ID,
		"url", article.Source)
	return nil
}

// GetArticle retrieves an article from storage.
func (s *Storage) GetArticle(ctx context.Context, id string) (*models.Article, error) {
	if id == "" {
		return nil, fmt.Errorf("article ID is empty")
	}

	article := &models.Article{}
	if err := s.storage.GetDocument(ctx, s.indexName, id, article); err != nil {
		s.logger.Error("Failed to get article",
			"error", err,
			"articleID", id)
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	s.logger.Debug("Retrieved article",
		"articleID", id)
	return article, nil
}

// ListArticles lists articles matching the query.
func (s *Storage) ListArticles(ctx context.Context, query string) ([]*models.Article, error) {
	// Create a search query
	searchQuery := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title^2", "body", "description"},
			},
		},
		"size": 100, // Default size of 100
	}

	// Execute the search
	results, err := s.storage.Search(ctx, s.indexName, searchQuery)
	if err != nil {
		s.logger.Error("Failed to list articles",
			"error", err,
			"query", query)
		return nil, fmt.Errorf("failed to list articles: %w", err)
	}

	// Convert results to articles
	articles := make([]*models.Article, 0, len(results))
	for _, result := range results {
		if article, ok := result.(*models.Article); ok {
			articles = append(articles, article)
		} else {
			// Try to unmarshal if it's a map
			if m, ok := result.(map[string]any); ok {
				article := &models.Article{}
				if data, err := json.Marshal(m); err == nil {
					if err := json.Unmarshal(data, article); err == nil {
						articles = append(articles, article)
					}
				}
			}
		}
	}

	s.logger.Debug("Listed articles",
		"query", query,
		"count", len(articles))
	return articles, nil
}

// Ensure Storage implements ArticleStorage interface
var _ ArticleStorage = (*Storage)(nil)
