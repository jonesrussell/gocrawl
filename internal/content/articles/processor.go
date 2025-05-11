// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"io"

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
func NewProcessor(
	logger logger.Interface,
	service Interface,
	validator content.JobValidator,
	storage types.Interface,
	indexName string,
	articleChannel chan *models.Article,
	articleIndexer processor.Processor,
	pageIndexer processor.Processor,
) *ArticleProcessor {
	return &ArticleProcessor{
		logger:         logger,
		service:        service,
		validator:      validator,
		storage:        storage,
		indexName:      indexName,
		articleChannel: articleChannel,
		articleIndexer: articleIndexer,
		pageIndexer:    pageIndexer,
	}
}

// Process implements the common.Processor interface.
func (p *ArticleProcessor) Process(ctx context.Context, content any) error {
	// TODO: Implement article processing
	return errors.New("not implemented")
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
			return errors.New("invalid item type: expected article")
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
	return nil, errors.New("unsupported content type")
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

// Get retrieves an article by its ID
func (p *ArticleProcessor) Get(ctx context.Context, id string) (*models.Article, error) {
	// TODO: Implement article retrieval
	return nil, errors.New("not implemented")
}

// GetByURL retrieves an article by its URL
func (p *ArticleProcessor) GetByURL(ctx context.Context, url string) (string, error) {
	// TODO: Implement article retrieval by URL
	return "", errors.New("not implemented")
}

// Validate validates a job
func (p *ArticleProcessor) Validate(job *content.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if len(job.Items) == 0 {
		return errors.New("job must have at least one item")
	}
	return nil
}
