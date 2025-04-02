// Package article provides functionality for processing and managing articles.
package article

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ArticleProcessor handles article content processing.
type ArticleProcessor struct {
	// Logger for article processing operations
	Logger common.Logger
	// ArticleService for article operations
	ArticleService Interface
	// Storage for article persistence
	Storage storagetypes.Interface
	// IndexName is the name of the article index
	IndexName string
	// ArticleChan is the channel for sending processed articles
	ArticleChan chan *models.Article
	// metrics holds processing metrics
	metrics *common.Metrics
}

// NewArticleProcessor creates a new article processor.
func NewArticleProcessor(p ProcessorParams) *ArticleProcessor {
	return &ArticleProcessor{
		Logger:         p.Logger,
		ArticleService: p.Service,
		Storage:        p.Storage,
		IndexName:      p.IndexName,
		ArticleChan:    p.ArticleChan,
		metrics:        &common.Metrics{},
	}
}

// Start implements common.Processor.Start.
func (p *ArticleProcessor) Start(ctx context.Context) error {
	p.Logger.Info("Starting article processor",
		"component", "article/processor")
	return nil
}

// Stop implements common.Processor.Stop.
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	p.Logger.Info("Stopping article processor",
		"component", "article/processor")
	return nil
}

// ProcessJob processes a job and its items.
func (p *ArticleProcessor) ProcessJob(ctx context.Context, job *common.Job) {
	start := time.Now()
	defer func() {
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		p.Logger.Warn("Job processing cancelled",
			"job_id", job.ID,
			"error", ctx.Err(),
		)
		p.metrics.ErrorCount++
		return
	default:
		// Process the job
		p.Logger.Info("Processing job",
			"job_id", job.ID,
		)

		// TODO: Implement job processing logic
		// This would typically involve:
		// 1. Fetching items associated with the job
		// 2. Processing each item
		// 3. Updating job status
		// 4. Handling errors and retries

		p.metrics.ProcessedCount++
	}
}

// ProcessHTML processes HTML content from a source.
func (p *ArticleProcessor) ProcessHTML(e *colly.HTMLElement) error {
	start := time.Now()
	defer func() {
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	// Extract article data from HTML element
	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article found in HTML element",
			"component", "article/processor",
			"url", e.Request.URL.String())
		return nil
	}

	// Process the article
	if err := p.ArticleService.Process(article); err != nil {
		p.Logger.Error("Failed to process article",
			"component", "article/processor",
			"articleID", article.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	// Send to channel if available
	if p.ArticleChan != nil {
		p.ArticleChan <- article
	}

	p.metrics.ProcessedCount++
	p.metrics.LastProcessedTime = time.Now()

	return nil
}

// GetMetrics returns the current processing metrics.
func (p *ArticleProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// ProcessHTMLElement handles article extraction from HTML elements
func (p *ArticleProcessor) ProcessHTMLElement(e *colly.HTMLElement) {
	p.Logger.Debug("Processing article from HTML",
		"component", "article/processor",
		"url", e.Request.URL.String())

	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article extracted",
			"component", "article/processor",
			"url", e.Request.URL.String())
		return
	}

	p.Logger.Debug("Article extracted",
		"component", "article/processor",
		"url", e.Request.URL.String(),
		"title", article.Title)

	if err := p.ArticleService.Process(article); err != nil {
		p.Logger.Error("Failed to process article",
			"component", "article/processor",
			"articleID", article.ID,
			"error", err)
	}
}

// ProcessContent implements the collector.Processor interface
func (p *ArticleProcessor) ProcessContent(e *colly.HTMLElement) {
	// Skip content pages - we only process articles
	p.Logger.Debug("Skipping content page in article processor",
		"component", "article/processor",
		"url", e.Request.URL.String())
}

// Process implements common.Processor.Process.
func (p *ArticleProcessor) Process(ctx context.Context, data interface{}) error {
	article, ok := data.(*models.Article)
	if !ok {
		return fmt.Errorf("invalid data type: expected *models.Article, got %T", data)
	}

	// Process the article using the ArticleService
	return p.ArticleService.Process(article)
}

// Ensure ArticleProcessor implements common.Processor
var _ common.Processor = (*ArticleProcessor)(nil)
