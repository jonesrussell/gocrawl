package search

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
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
	Logger   common.Logger
	Config   *config.Config
	Options  storage.Options
	Storage  storage.Interface
}

// NewSearchService creates a new instance of the search service
func NewSearchService(
	esClient *elasticsearch.Client,
	cfg *config.Config,
	log common.Logger,
) *Service {
	opts := storage.NewOptionsFromConfig(cfg)
	return &Service{
		ESClient: esClient,
		Logger:   log,
		Config:   cfg,
		Options:  opts,
	}
}

// SearchContent performs a search query
func (s *Service) SearchContent(ctx context.Context, query string, _ string, size int) ([]Result, error) {
	var storageInstance storage.Interface
	var err error

	if s.Storage != nil {
		storageInstance = s.Storage
	} else {
		storageInstance, err = storage.NewStorage(s.ESClient, s.Options, s.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}
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
	}

	return searchResults, nil
}
