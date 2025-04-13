// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ArticleProcessor implements the common.Processor interface for articles.
type ArticleProcessor struct {
	logger         logger.Interface
	service        Interface
	jobService     common.JobService
	storage        types.Interface
	indexName      string
	articleChannel chan *models.Article
}

// NewProcessor creates a new article processor.
func NewProcessor(p ProcessorParams) *ArticleProcessor {
	return &ArticleProcessor{
		logger:         p.Logger,
		service:        p.Service,
		jobService:     p.JobService,
		storage:        p.Storage,
		indexName:      p.IndexName,
		articleChannel: p.ArticleChannel,
	}
}

// Process implements the common.Processor interface.
func (p *ArticleProcessor) Process(ctx context.Context, content any) error {
	// Check if content is a job
	job, ok := content.(*jobtypes.Job)
	if !ok {
		return fmt.Errorf("invalid content type: expected *jobtypes.Job, got %T", content)
	}

	// Create a new collector for this job
	c := colly.NewCollector()

	// Configure the collector
	c.OnHTML("article", func(e *colly.HTMLElement) {
		p.service.Process(e)
	})

	// Visit the URL
	if err := c.Visit(job.URL); err != nil {
		return fmt.Errorf("failed to visit URL: %w", err)
	}

	// Send the processed article to the channel
	if p.articleChannel != nil {
		article := &models.Article{
			Source:    job.URL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		p.articleChannel <- article
	}

	return nil
}

// ContentType returns the type of content this processor handles.
func (p *ArticleProcessor) ContentType() common.ContentType {
	return common.ContentTypeArticle
}

// CanProcess implements the common.Processor interface.
func (p *ArticleProcessor) CanProcess(contentType any) bool {
	ct, ok := contentType.(common.ContentType)
	return ok && ct == common.ContentTypeArticle
}

// ParseHTML implements the common.Processor interface.
func (p *ArticleProcessor) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// ExtractLinks implements the common.Processor interface.
func (p *ArticleProcessor) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// ExtractContent implements the common.Processor interface.
func (p *ArticleProcessor) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
}

// ProcessJob processes a job and its items.
func (p *ArticleProcessor) ProcessJob(ctx context.Context, job *common.Job) error {
	return p.Process(ctx, job)
}

// ValidateJob implements the common.Processor interface.
func (p *ArticleProcessor) ValidateJob(job *jobtypes.Job) error {
	if job == nil {
		return errors.New("job is nil")
	}
	if job.URL == "" {
		return errors.New("job URL is empty")
	}
	return nil
}

// RegisterProcessor registers a new content processor.
func (p *ArticleProcessor) RegisterProcessor(processor common.ContentProcessor) {
	// Not implemented - we only handle article processing
}

// GetProcessor returns a processor for the given content type.
func (p *ArticleProcessor) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	if contentType == common.ContentTypeArticle {
		return p, nil
	}
	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

// ProcessContent implements the common.Processor interface.
func (p *ArticleProcessor) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	if contentType != common.ContentTypeArticle {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	job, ok := content.(common.Job)
	if !ok {
		return errors.New("invalid content type: expected common.Job")
	}
	return p.Process(ctx, job)
}

// Start initializes the processor.
func (p *ArticleProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop cleans up the processor.
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	return p.Close()
}

// Close implements the common.Processor interface.
func (p *ArticleProcessor) Close() error {
	if p.articleChannel != nil {
		close(p.articleChannel)
	}
	return nil
}
