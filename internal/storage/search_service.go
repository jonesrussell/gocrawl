package storage

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/models"
)

// SearchServiceInterface defines the methods for the SearchService
type SearchServiceInterface interface {
	SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error)
}

// SearchService implements the SearchServiceInterface
type SearchService struct {
	storage Interface
}

// NewSearchService creates a new SearchService instance
func NewSearchService(storage Interface) SearchServiceInterface {
	return &SearchService{
		storage: storage,
	}
}

// SearchArticles searches for articles based on a query
func (s *SearchService) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	// Ensure the connection is valid
	if err := s.storage.TestConnection(ctx); err != nil {
		return nil, err
	}

	// Check if the index exists
	exists, err := s.storage.IndexExists(ctx, "articles")
	if err != nil || !exists {
		return nil, err
	}

	// Perform the search
	articles, err := s.storage.SearchArticles(ctx, query, size)
	if err != nil {
		return nil, err
	}

	return articles, nil
}
