// Package article provides functionality for processing and managing articles.
package article

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ProcessorParams defines the parameters for creating a new article processor.
type ProcessorParams struct {
	// Logger for article processing operations
	Logger logger.Interface
	// Service for article operations
	Service Interface
	// Storage for article persistence
	Storage storagetypes.Interface
	// IndexName is the name of the article index
	IndexName string
	// ArticleChan is the channel for sending processed articles
	ArticleChan chan *models.Article
}

// ArticleProcessor handles article content processing.
type ArticleProcessor struct {
	// Logger for article processing operations
	Logger logger.Interface
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
	if p.Logger == nil {
		panic("logger is required")
	}
	if p.Service == nil {
		panic("article service is required")
	}
	if p.Storage == nil {
		panic("storage is required")
	}
	if p.IndexName == "" {
		panic("index name is required")
	}

	return &ArticleProcessor{
		Logger:         p.Logger,
		ArticleService: p.Service,
		Storage:        p.Storage,
		IndexName:      p.IndexName,
		ArticleChan:    p.ArticleChan,
		metrics:        &common.Metrics{},
	}
}

// ContentType implements ContentProcessor.ContentType
func (p *ArticleProcessor) ContentType() common.ContentType {
	return common.ContentTypeArticle
}

// CanProcess checks if this processor can handle the given content
func (p *ArticleProcessor) CanProcess(content any) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// Process processes the content and returns the processed result
func (p *ArticleProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}
	return p.ProcessHTML(ctx, e)
}

// ProcessHTML implements HTMLProcessor.ProcessHTML
func (p *ArticleProcessor) ProcessHTML(ctx context.Context, e *colly.HTMLElement) error {
	if p == nil {
		return errors.New("article processor is nil")
	}
	if p.ArticleService == nil {
		return errors.New("article service is nil")
	}
	if p.Logger == nil {
		return errors.New("logger is nil")
	}

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

// ExtractContent implements HTMLProcessor.ExtractContent
func (p *ArticleProcessor) ExtractContent() (string, error) {
	// For article processing, we don't need to extract raw content
	// as we process the structured article data instead
	return "", nil
}

// ProcessJob implements JobProcessor.ProcessJob
func (p *ArticleProcessor) ProcessJob(ctx context.Context, job *common.Job) error {
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
		return ctx.Err()
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
		return nil
	}
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

// Start implements Processor.Start
func (p *ArticleProcessor) Start(ctx context.Context) error {
	p.Logger.Info("Starting article processor")
	return nil
}

// Stop implements Processor.Stop
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	p.Logger.Info("Stopping article processor")
	return nil
}

// ExtractLinks implements HTMLProcessor.ExtractLinks
func (p *ArticleProcessor) ExtractLinks() ([]string, error) {
	// For article processing, we don't need to extract links
	// as we process the structured article data instead
	return nil, nil
}

// ValidateJob implements JobProcessor.ValidateJob
func (p *ArticleProcessor) ValidateJob(job *common.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if job.ID == "" {
		return errors.New("job ID cannot be empty")
	}
	return nil
}

// ParseHTML implements HTMLProcessor.ParseHTML
func (p *ArticleProcessor) ParseHTML(r io.Reader) error {
	// For article processing, we don't need to parse raw HTML
	// as we process the structured article data instead
	return nil
}

// Ensure ArticleProcessor implements required interfaces
var (
	_ common.ContentProcessor = (*ArticleProcessor)(nil)
	_ common.HTMLProcessor    = (*ArticleProcessor)(nil)
	_ common.JobProcessor     = (*ArticleProcessor)(nil)
)
