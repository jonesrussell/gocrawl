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
	"github.com/jonesrussell/gocrawl/internal/common/contenttype"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ArticleProcessor implements the common.Processor interface for articles.
type ArticleProcessor struct {
	logger         logger.Interface
	service        Interface
	validator      jobtypes.JobValidator
	storage        types.Interface
	indexName      string
	articleChannel chan *models.Article
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
func (p *ArticleProcessor) ValidateJob(job *jobtypes.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if job.Type != contenttype.Article {
		return fmt.Errorf("invalid job type: expected %s, got %s", contenttype.Article, job.Type)
	}
	return nil
}

// RegisterProcessor registers a new content processor.
func (p *ArticleProcessor) RegisterProcessor(processor common.Processor) {
	// Not implemented - we only handle article processing
}

// GetProcessor returns the processor for the given content type.
func (p *ArticleProcessor) GetProcessor(contentType contenttype.Type) (common.Processor, error) {
	if contentType == contenttype.Article {
		return p, nil
	}
	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

// Start implements the common.Processor interface.
func (p *ArticleProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements the common.Processor interface.
func (p *ArticleProcessor) Stop(ctx context.Context) error {
	return nil
}

// Close cleans up resources used by the processor.
func (p *ArticleProcessor) Close() error {
	if p.articleChannel != nil {
		close(p.articleChannel)
	}
	return nil
}
