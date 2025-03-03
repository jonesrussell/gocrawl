package search

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Result represents a search result
type Result struct {
	URL     string
	Content string
}

// Service represents the search service
type Service struct {
	ESClient *elasticsearch.Client
	Logger   logger.Interface
	Config   *config.Config
}

// NewSearchService creates a new instance of the search service
func NewSearchService(esClient *elasticsearch.Client, cfg *config.Config, log logger.Interface) *Service {
	return &Service{
		ESClient: esClient,
		Logger:   log,
		Config:   cfg,
	}
}

// SearchContent performs a search query
func (s *Service) SearchContent(ctx context.Context, query string, _ string, size int) ([]Result, error) {
	storageInstance, err := storage.NewStorage(s.ESClient, s.Config, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	results, err := storageInstance.SearchArticles(ctx, query, size)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	searchResults := make([]Result, len(results))
	for i, article := range results {
		searchResults[i] = Result{
			URL:     article.Source,
			Content: article.Body,
		}
		s.Logger.Info("URL: %s, Content: %s", article.Source, article.Body)
	}

	return searchResults, nil
}
