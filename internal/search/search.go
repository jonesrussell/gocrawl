package search

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// SearchResult represents a search result
type SearchResult struct {
	URL     string
	Content string
}

// SearchService represents the search service
type SearchService struct {
	ESClient *elasticsearch.Client
	Logger   logger.Interface
}

// NewSearchService creates a new instance of the search service
func NewSearchService(esClient *elasticsearch.Client, log logger.Interface) *SearchService {
	return &SearchService{
		ESClient: esClient,
		Logger:   log,
	}
}

// SearchContent performs a search query
func (s *SearchService) SearchContent(ctx context.Context, query string, index string, size int) ([]SearchResult, error) {
	storageInstance, err := storage.NewStorage(s.ESClient, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	results, err := storageInstance.SearchArticles(ctx, query, size)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	searchResults := make([]SearchResult, len(results))
	for i, article := range results {
		searchResults[i] = SearchResult{
			URL:     article.Source,
			Content: article.Body,
		}
		s.Logger.Info("URL: %s, Content: %s", article.Source, article.Body)
	}

	return searchResults, nil
}
