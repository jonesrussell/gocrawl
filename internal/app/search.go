package app

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/zap"
)

// SearchResult represents a search result
type SearchResult struct {
	URL     string
	Content string
}

// SearchContent performs a search query
func SearchContent(ctx context.Context, query string, index string, size int) ([]SearchResult, error) {
	log, err := logger.NewDevelopmentLogger(logger.Params{
		Debug:  true,
		Level:  zap.InfoLevel,
		AppEnv: "development",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	storageResult, err := storage.NewStorage(&config.Config{
		Crawler: config.CrawlerConfig{
			IndexName: index,
		},
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200", // Consider making this configurable
		},
	}, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	results, err := storageResult.Storage.SearchArticles(ctx, query, size)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	searchResults := make([]SearchResult, len(results))
	for i, article := range results {
		searchResults[i] = SearchResult{
			URL:     article.Source,
			Content: article.Body,
		}
	}

	return searchResults, nil
}
