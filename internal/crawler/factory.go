// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"errors"

	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ProcessorFactory creates content processors for different content types.
type ProcessorFactory interface {
	// CreateProcessor creates a new processor for the given content type.
	CreateProcessor(contentType contenttype.Type) (content.Processor, error)
}

// DefaultProcessorFactory implements ProcessorFactory with default processors.
type DefaultProcessorFactory struct {
	logger     logger.Interface
	storage    types.Interface
	indexName  string
	processors map[contenttype.Type]content.Processor
}

// NewProcessorFactory creates a new processor factory.
func NewProcessorFactory(
	log logger.Interface,
	storage types.Interface,
	indexName string,
) ProcessorFactory {
	return &DefaultProcessorFactory{
		logger:     log,
		storage:    storage,
		indexName:  indexName,
		processors: make(map[contenttype.Type]content.Processor),
	}
}

// CreateProcessor implements ProcessorFactory.
func (f *DefaultProcessorFactory) CreateProcessor(contentType contenttype.Type) (content.Processor, error) {
	// Check if we already have a processor for this type
	if processor, ok := f.processors[contentType]; ok {
		return processor, nil
	}

	// Create a new processor based on the content type
	var processor content.Processor

	switch contentType {
	case contenttype.Article:
		processor = f.createArticleProcessor()
	case contenttype.Page:
		processor = f.createPageProcessor()
	case contenttype.Video:
		return nil, errors.New("video processing not implemented")
	case contenttype.Image:
		return nil, errors.New("image processing not implemented")
	case contenttype.HTML:
		return nil, errors.New("HTML processing not implemented")
	case contenttype.Job:
		return nil, errors.New("job processing not implemented")
	default:
		return nil, errors.New("unsupported content type")
	}

	// Cache the processor for future use
	f.processors[contentType] = processor

	return processor, nil
}

// createArticleProcessor creates a new article processor
func (f *DefaultProcessorFactory) createArticleProcessor() content.Processor {
	// Create a simple job validator
	validator := &struct {
		content.JobValidator
	}{
		JobValidator: content.JobValidatorFunc(func(job *content.Job) error {
			if job == nil {
				return errors.New("job cannot be nil")
			}
			if job.URL == "" {
				return errors.New("job URL cannot be empty")
			}
			return nil
		}),
	}

	// Create article service
	articleService := articles.NewContentService(f.logger, f.storage, f.indexName)

	return articles.NewProcessor(
		f.logger,
		articleService,
		validator,
		f.storage,
		"articles",
		make(chan *domain.Article, ArticleChannelBufferSize),
		nil,
		nil,
	)
}

// createPageProcessor creates a new page processor
func (f *DefaultProcessorFactory) createPageProcessor() content.Processor {
	// Create a simple job validator
	validator := &struct {
		content.JobValidator
	}{
		JobValidator: content.JobValidatorFunc(func(job *content.Job) error {
			if job == nil {
				return errors.New("job cannot be nil")
			}
			if job.URL == "" {
				return errors.New("job URL cannot be empty")
			}
			return nil
		}),
	}

	// Create page service
	pageService := page.NewContentService(f.logger, f.storage, f.indexName)

	return page.NewPageProcessor(
		f.logger,
		pageService,
		validator,
		f.storage,
		"pages",
		make(chan *domain.Page, DefaultChannelBufferSize),
	)
}
