package app

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

// SearchContent performs a search query
func SearchContent(ctx context.Context, query string, index string, size int) ([]SearchResult, error) {
	log, err := logger.NewDevelopmentLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"}, // Consider making this configurable
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	storageInstance, err := storage.NewStorage(esClient)
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
		log.Info("URL: %s, Content: %s", article.Source, article.Body)
	}

	return searchResults, nil
}
