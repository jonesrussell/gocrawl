// Package article provides functionality for processing and managing articles.
package article

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// Interface defines the interface for article processing operations.
type Interface interface {
	// ExtractArticle extracts article data from an HTML element.
	ExtractArticle(e *colly.HTMLElement) *models.Article
	// Process processes an article.
	Process(article *models.Article) error
	// ProcessJob processes a job and its items.
	ProcessJob(ctx context.Context, job *common.Job)
	// ProcessHTML processes HTML content from a source.
	ProcessHTML(e *colly.HTMLElement) error
	// GetMetrics returns the current processing metrics.
	GetMetrics() *common.Metrics
}
