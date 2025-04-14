// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/processor"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ArticleProcessor implements the common.Processor interface for articles.
type ArticleProcessor struct {
	logger         logger.Interface
	service        Interface
	validator      content.JobValidator
	storage        types.Interface
	indexName      string
	articleChannel chan *models.Article
	articleIndexer processor.Processor
	pageIndexer    processor.Processor
}

// NewProcessor creates a new article processor.
func NewProcessor(p ProcessorParams) *ArticleProcessor {
	return &ArticleProcessor{
		logger:         p.Logger,
		service:        p.Service,
		validator:      p.Validator,
		storage:        p.Storage,
		indexName:      p.IndexName,
		articleChannel: p.ArticleChannel,
		articleIndexer: p.ArticleIndexer,
		pageIndexer:    p.PageIndexer,
	}
}

// Process implements the common.Processor interface.
func (p *ArticleProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Process the article
	if err := p.service.Process(e); err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}

	// Send the processed article to the channel
	if p.articleChannel != nil {
		article := &models.Article{
			Source:    e.Request.URL.String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		p.articleChannel <- article
	}

	return nil
}

// ContentType implements the common.Processor interface.
func (p *ArticleProcessor) ContentType() contenttype.Type {
	return contenttype.Article
}

// CanProcess implements the common.Processor interface.
func (p *ArticleProcessor) CanProcess(contentType contenttype.Type) bool {
	return contentType == contenttype.Article
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

// ValidateJob implements the common.Processor interface.
func (p *ArticleProcessor) ValidateJob(job *content.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if len(job.Items) == 0 {
		return errors.New("job must have at least one item")
	}
	for _, item := range job.Items {
		if item.Type != content.Article {
			return fmt.Errorf("invalid item type: expected %s, got %s", content.Article, item.Type)
		}
	}
	return nil
}

// RegisterProcessor implements content.ProcessorRegistry
func (p *ArticleProcessor) RegisterProcessor(processor content.ContentProcessor) {
	// Not implemented - we only handle article processing
}

// GetProcessor implements content.ProcessorRegistry
func (p *ArticleProcessor) GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error) {
	if contentType == contenttype.Article {
		return &articleContentProcessor{p}, nil
	}
	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

// articleContentProcessor wraps ArticleProcessor to implement content.ContentProcessor
type articleContentProcessor struct {
	*ArticleProcessor
}

// Process implements content.ContentProcessor
func (p *articleContentProcessor) Process(ctx context.Context, content any) error {
	return p.ArticleProcessor.Process(ctx, content)
}

// ContentType implements content.ContentProcessor
func (p *articleContentProcessor) ContentType() contenttype.Type {
	return p.ArticleProcessor.ContentType()
}

// CanProcess implements content.ContentProcessor
func (p *articleContentProcessor) CanProcess(content contenttype.Type) bool {
	return p.ArticleProcessor.CanProcess(content)
}

// ValidateJob implements content.ContentProcessor
func (p *articleContentProcessor) ValidateJob(job *content.Job) error {
	return p.ArticleProcessor.ValidateJob(job)
}

// Start implements content.Processor
func (p *ArticleProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements content.Processor
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	return p.Close()
}

// Close cleans up resources used by the processor.
func (p *ArticleProcessor) Close() error {
	if p.articleChannel != nil {
		close(p.articleChannel)
	}
	return nil
}

// ProcessContent implements content.ProcessorRegistry
func (p *ArticleProcessor) ProcessContent(ctx context.Context, contentType contenttype.Type, content any) error {
	processor, err := p.GetProcessor(contentType)
	if err != nil {
		return err
	}
	return processor.Process(ctx, content)
}
