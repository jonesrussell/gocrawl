package content

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/pkg/collector"
)

// ContentProcessor handles the processing of non-article content
type ContentProcessor struct {
	service   Interface
	storage   common.Storage
	logger    logger.Interface
	indexName string
}

// NewProcessor creates a new content processor instance.
func NewProcessor(
	service Interface,
	storage common.Storage,
	logger logger.Interface,
	indexName string,
) *ContentProcessor {
	return &ContentProcessor{
		service:   service,
		storage:   storage,
		logger:    logger,
		indexName: indexName,
	}
}

// Process implements the collector.Processor interface
func (p *ContentProcessor) Process(e *colly.HTMLElement) error {
	p.logger.Debug("Processing content",
		"component", "content/processor",
		"url", e.Request.URL.String(),
		"index", p.indexName)

	content := p.service.ExtractContent(e)
	if content == nil {
		p.logger.Debug("No content extracted",
			"component", "content/processor",
			"url", e.Request.URL.String())
		return nil
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
		return err
	}

	p.logger.Debug("Content processed successfully",
		"component", "content/processor",
		"url", content.URL,
		"id", content.ID,
		"type", content.Type,
		"index", p.indexName)

	return nil
}

// Ensure ContentProcessor implements collector.Processor
var _ collector.Processor = (*ContentProcessor)(nil)
