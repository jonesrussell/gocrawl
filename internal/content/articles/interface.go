// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/domain"
)

// Interface defines the interface for processing articles.
// It handles the extraction and processing of article content from web pages.
type Interface interface {
	// Process handles the processing of article content.
	// It takes a colly.HTMLElement and processes the article found within it.
	Process(e *colly.HTMLElement) error

	// ProcessArticle processes an article and returns any errors
	ProcessArticle(ctx context.Context, article *domain.Article) error

	// Get retrieves an article by its ID
	Get(ctx context.Context, id string) (*domain.Article, error)

	// List retrieves a list of articles based on the provided query
	List(ctx context.Context, query map[string]any) ([]*domain.Article, error)

	// Delete removes an article by its ID
	Delete(ctx context.Context, id string) error

	// Update updates an existing article
	Update(ctx context.Context, article *domain.Article) error

	// Create creates a new article
	Create(ctx context.Context, article *domain.Article) error
}
