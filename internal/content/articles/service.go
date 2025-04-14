// Package articles provides functionality for processing and managing article content.
package articles

import (
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentService implements the Interface for article processing.
type ContentService struct {
	logger    logger.Interface
	storage   types.Interface
	indexName string
}

// NewContentService creates a new article content service.
func NewContentService(p ServiceParams) Interface {
	return &ContentService{
		logger:    p.Logger,
		storage:   p.Storage,
		indexName: p.IndexName,
	}
}

// Process implements the Interface.
func (s *ContentService) Process(e *colly.HTMLElement) error {
	if e == nil {
		return errors.New("nil HTML element")
	}

	s.logger.Debug("Processing article",
		"url", e.Request.URL.String())

	// Extract article data from the HTML element
	// This is a placeholder - implement actual extraction logic
	s.logger.Debug("Article processed",
		"url", e.Request.URL.String())

	return nil
}
