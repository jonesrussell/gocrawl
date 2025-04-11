// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// CrawlerInterface defines the core functionality of a crawler.
type CrawlerInterface interface {
	// Start begins crawling from the given source.
	Start(ctx context.Context, sourceName string) error
	// Stop gracefully stops the crawler.
	Stop(ctx context.Context) error
	// Subscribe adds a handler for crawler events.
	Subscribe(handler EventHandler)
	// GetMetrics returns the current crawler metrics.
	GetMetrics() *common.Metrics
}

// Config defines the configuration for a crawler.
type Config struct {
	// MaxDepth is the maximum depth to crawl.
	MaxDepth int
	// RateLimit is the delay between requests.
	RateLimit time.Duration
	// Parallelism is the number of concurrent requests.
	Parallelism int
	// AllowedDomains are the domains that can be crawled.
	AllowedDomains []string
	// UserAgent is the user agent string to use.
	UserAgent string
}

// CrawlerState manages the runtime state of a crawler.
type CrawlerState interface {
	// IsRunning returns whether the crawler is running.
	IsRunning() bool
	// StartTime returns when the crawler started.
	StartTime() time.Time
	// CurrentSource returns the current source being crawled.
	CurrentSource() string
	// Context returns the crawler's context.
	Context() context.Context
	// Cancel cancels the crawler's context.
	Cancel()
}

// CrawlerMetrics tracks crawler statistics.
type CrawlerMetrics interface {
	// IncrementProcessed increments the processed count.
	IncrementProcessed()
	// IncrementError increments the error count.
	IncrementError()
	// GetProcessedCount returns the number of processed items.
	GetProcessedCount() int64
	// GetErrorCount returns the number of errors.
	GetErrorCount() int64
	// GetStartTime returns when tracking started.
	GetStartTime() time.Time
}

// ContentProcessor handles content processing.
type ContentProcessor interface {
	// ProcessHTML processes HTML content.
	ProcessHTML(ctx context.Context, element *colly.HTMLElement) error
	// CanProcess returns whether the processor can handle the content.
	CanProcess(contentType string) bool
	// ContentType returns the content type this processor handles.
	ContentType() string
}

// ArticleStorage handles data persistence.
type ArticleStorage interface {
	// SaveArticle saves an article.
	SaveArticle(ctx context.Context, article *models.Article) error
	// GetArticle retrieves an article.
	GetArticle(ctx context.Context, id string) (*models.Article, error)
	// ListArticles lists articles matching the query.
	ListArticles(ctx context.Context, query string) ([]*models.Article, error)
}

// EventHandler handles crawler events.
type EventHandler interface {
	// HandleArticle handles a discovered article.
	HandleArticle(ctx context.Context, article *models.Article) error
	// HandleError handles a crawler error.
	HandleError(ctx context.Context, err error) error
	// HandleStart handles crawler start.
	HandleStart(ctx context.Context) error
	// HandleStop handles crawler stop.
	HandleStop(ctx context.Context) error
}

// EventBus manages event distribution.
type EventBus interface {
	// Subscribe adds an event handler.
	Subscribe(handler EventHandler)
	// PublishArticle publishes an article event.
	PublishArticle(ctx context.Context, article *models.Article) error
	// PublishError publishes an error event.
	PublishError(ctx context.Context, err error) error
	// PublishStart publishes a start event.
	PublishStart(ctx context.Context) error
	// PublishStop publishes a stop event.
	PublishStop(ctx context.Context) error
}
