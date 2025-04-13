// Package article provides functionality for processing and managing articles.
package article

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	// JobService for job operations
	JobService common.JobService
	// Storage for article persistence
	Storage storagetypes.Interface
	// IndexName is the name of the article index
	IndexName string
	// ArticleChan is the channel for sending processed articles
	ArticleChan chan *models.Article
}

// ArticleProcessor implements the Processor interface for article processing.
type ArticleProcessor struct {
	ArticleService Interface
	Logger         logger.Interface
	metrics        *common.Metrics
	ArticleChan    chan *models.Article
	JobService     common.JobService
	Storage        storagetypes.Interface
	IndexName      string
}

// NewArticleProcessor creates a new article processor.
func NewArticleProcessor(p ProcessorParams) *ArticleProcessor {
	if p.Logger == nil {
		panic("logger is required")
	}
	if p.Service == nil {
		panic("article service is required")
	}
	if p.JobService == nil {
		panic("job service is required")
	}
	if p.Storage == nil {
		panic("storage is required")
	}
	if p.IndexName == "" {
		panic("index name is required")
	}

	return &ArticleProcessor{
		ArticleService: p.Service,
		Logger:         p.Logger,
		metrics:        &common.Metrics{},
		ArticleChan:    p.ArticleChan,
		JobService:     p.JobService,
		Storage:        p.Storage,
		IndexName:      p.IndexName,
	}
}

// ContentType returns the type of content this processor can handle.
func (p *ArticleProcessor) ContentType() common.ContentType {
	return common.ContentTypeArticle
}

// CanProcess checks if the processor can handle the given content.
func (p *ArticleProcessor) CanProcess(content any) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// Process handles the content processing.
func (p *ArticleProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}
	return p.ProcessHTML(e)
}

// ParseHTML parses HTML content from a reader.
func (p *ArticleProcessor) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// ExtractLinks extracts links from the parsed HTML.
func (p *ArticleProcessor) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// ExtractContent extracts the main content from the parsed HTML.
func (p *ArticleProcessor) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
}

// ProcessJob processes a job and its items.
func (p *ArticleProcessor) ProcessJob(ctx context.Context, job *common.Job) error {
	if job == nil {
		return errors.New("job is nil")
	}

	// Get items for this job
	items, err := p.JobService.GetItems(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get job items: %w", err)
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			article := &models.Article{
				Source: item.URL,
			}
			if processErr := p.ProcessArticle(article); processErr != nil {
				p.Logger.Error("Failed to process article",
					"error", processErr)
			}
		}
	}

	return nil
}

// ValidateJob validates a job before processing.
func (p *ArticleProcessor) ValidateJob(job *common.Job) error {
	if job == nil {
		return errors.New("job is nil")
	}
	if job.ID == "" {
		return errors.New("job ID is empty")
	}
	return nil
}

// RegisterProcessor registers a new content processor.
func (p *ArticleProcessor) RegisterProcessor(processor common.ContentProcessor) {
	// Not implemented
}

// GetProcessor returns a processor for the given content type.
func (p *ArticleProcessor) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	if contentType == common.ContentTypeArticle {
		return p, nil
	}
	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

// ProcessContent processes content using the appropriate processor.
func (p *ArticleProcessor) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	if contentType != common.ContentTypeArticle {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	return p.Process(ctx, content)
}

// Start initializes the processor.
func (p *ArticleProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop cleans up the processor.
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	return nil
}

// ProcessArticle processes an article.
func (p *ArticleProcessor) ProcessArticle(article *models.Article) error {
	if article == nil {
		return errors.New("article is nil")
	}

	// Extract article data
	extracted := p.ArticleService.ExtractArticle(article.HTML)
	if extracted == nil {
		return errors.New("failed to extract article")
	}

	// Process the article
	if err := p.ArticleService.Process(extracted); err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}

	return nil
}

// ProcessHTML processes HTML content from a source.
func (p *ArticleProcessor) ProcessHTML(e *colly.HTMLElement) error {
	if e == nil {
		return errors.New("HTML element is nil")
	}

	p.Logger.Debug("Processing article from HTML",
		"component", "article/processor",
		"url", e.Request.URL.String())

	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article extracted",
			"component", "article/processor",
			"url", e.Request.URL.String())
		return errors.New("failed to extract article")
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
		return fmt.Errorf("failed to process article: %w", err)
	}

	return nil
}

// GetMetrics returns the current processing metrics.
func (p *ArticleProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// Ensure ArticleProcessor implements required interfaces
var (
	_ common.ContentProcessor = (*ArticleProcessor)(nil)
	_ common.HTMLProcessor    = (*ArticleProcessor)(nil)
	_ common.JobProcessor     = (*ArticleProcessor)(nil)
)
