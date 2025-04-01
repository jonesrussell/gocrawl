package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentService handles content extraction and processing.
type ContentService struct {
	logger common.Logger
}

// NewContentService creates a new content service.
func NewContentService(logger common.Logger) *ContentService {
	return &ContentService{
		logger: logger,
	}
}

// ContentProcessor implements the collector.Processor interface for general content.
type ContentProcessor struct {
	service   *ContentService
	storage   types.Interface
	logger    common.Logger
	indexName string
}

// Process implements the collector.Processor interface.
func (p *ContentProcessor) Process(e *colly.HTMLElement) error {
	// Extract content data
	content := &models.Content{
		ID:        uuid.New().String(),
		Title:     e.ChildText("h1"),
		Body:      e.ChildText("article"),
		Type:      "page",
		URL:       e.Request.URL.String(),
		CreatedAt: time.Now(),
	}

	// Store content
	if err := p.storage.IndexDocument(context.Background(), p.indexName, content.ID, content); err != nil {
		return fmt.Errorf("failed to store content: %w", err)
	}

	return nil
}
