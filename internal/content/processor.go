package content

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Processor handles the processing of non-article content
type Processor struct {
	service   Interface
	storage   types.Interface
	logger    logger.Interface
	indexName string
}

// NewProcessor creates a new content processor
func NewProcessor(service Interface, storage types.Interface, logger logger.Interface, indexName string) *Processor {
	return &Processor{
		service:   service,
		storage:   storage,
		logger:    logger,
		indexName: indexName,
	}
}

// Process implements the ContentProcessor interface
func (p *Processor) Process(e *colly.HTMLElement) {
	p.logger.Debug("Processing content",
		"component", "content/processor",
		"url", e.Request.URL.String(),
		"index", p.indexName)

	content := p.service.ExtractContent(e)
	if content == nil {
		p.logger.Debug("No content extracted",
			"component", "content/processor",
			"url", e.Request.URL.String())
		return
	}

	p.logger.Debug("Content extracted",
		"component", "content/processor",
		"url", content.URL,
		"id", content.ID,
		"type", content.Type,
		"title", content.Title)

	err := p.storage.IndexDocument(context.Background(), p.indexName, content.ID, content)
	if err != nil {
		p.logger.Error("Failed to index content",
			"component", "content/processor",
			"url", content.URL,
			"error", err,
		)
		return
	}

	p.logger.Debug("Content processed successfully",
		"component", "content/processor",
		"url", content.URL,
		"id", content.ID,
		"type", content.Type,
		"index", p.indexName)
}

// Ensure Processor implements models.ContentProcessor
var _ models.ContentProcessor = (*Processor)(nil)
