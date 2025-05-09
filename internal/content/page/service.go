// Package page provides functionality for processing and managing web pages.
package page

import (
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the contract for page processing services.
type Interface interface {
	Process(e *colly.HTMLElement) error
}

// ContentService implements the Interface for page processing.
type ContentService struct {
	logger    logger.Interface
	storage   types.Interface
	indexName string
}

// NewContentService creates a new page content service.
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

	s.logger.Debug("Processing page",
		"url", e.Request.URL.String())

	// Extract page data from the HTML element
	// This is a placeholder - implement actual extraction logic
	s.logger.Debug("Page processed",
		"url", e.Request.URL.String())

	return nil
}
