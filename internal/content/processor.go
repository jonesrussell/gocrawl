package content

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Processor handles the processing of non-article content
type Processor struct {
	service   Interface
	storage   storage.Interface
	logger    logger.Interface
	indexName string
}

// NewProcessor creates a new content processor
func NewProcessor(service Interface, storage storage.Interface, logger logger.Interface, indexName string) *Processor {
	return &Processor{
		service:   service,
		storage:   storage,
		logger:    logger,
		indexName: indexName,
	}
}

// ProcessContent implements the ContentProcessor interface
func (p *Processor) ProcessContent(e *colly.HTMLElement) {
	content := p.service.ExtractContent(e)
	if content == nil {
		return
	}

	ctx := context.Background()
	err := p.storage.IndexDocument(ctx, p.indexName, content.ID, content)
	if err != nil {
		p.logger.Error("Failed to index content",
			"url", content.URL,
			"error", err,
		)
		return
	}

	p.logger.Debug("Content indexed successfully",
		"url", content.URL,
		"id", content.ID,
		"type", content.Type,
		"index", p.indexName,
	)
}
