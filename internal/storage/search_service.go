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
	return s.storage.SearchArticles(ctx, query, size)
}
