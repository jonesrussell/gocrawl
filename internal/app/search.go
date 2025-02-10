package app

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// SearchResult represents a search result
type SearchResult struct {
	URL     string
	Content string
}

// SearchParams holds the parameters for search operations
type SearchParams struct {
	fx.In

	Config  *config.Config
	Storage storage.Storage
	Logger  logger.Interface
}

// SearchContent searches for content in the specified index
func SearchContent(ctx context.Context, query string, indexName string, size int) ([]SearchResult, error) {
	// Initialize the fx app with dependencies
	var searchResults []SearchResult
	var searchErr error

	app := fx.New(
		fx.Provide(
			config.LoadConfig,
			logger.NewCustomLogger,
			storage.NewStorage,
		),
		fx.Invoke(func(p SearchParams) {
			// Search for articles
			articles, err := p.Storage.SearchArticles(ctx, query, size)
			if err != nil {
				searchErr = fmt.Errorf("failed to search articles: %w", err)
				return
			}

			// Convert articles to search results
			results := make([]SearchResult, len(articles))
			for i, article := range articles {
				results[i] = SearchResult{
					URL:     article.Source,
					Content: article.Body,
				}
			}
			searchResults = results
		}),
	)

	if err := app.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start search: %w", err)
	}
	defer app.Stop(ctx)

	if searchErr != nil {
		return nil, searchErr
	}

	return searchResults, nil
}
