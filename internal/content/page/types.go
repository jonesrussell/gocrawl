// Package page provides functionality for processing and managing web pages.
package page

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// PageService defines the page service interface.
type PageService interface {
	// Process processes a page from an HTML element.
	Process(e *colly.HTMLElement) error
}

// ServiceParams contains the parameters for creating a new content service.
type ServiceParams struct {
	Logger    logger.Interface
	Storage   types.Interface
	IndexName string
}

// ModuleParams defines the parameters required for creating a page service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
type ModuleParams struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"pageIndexName"`
}

// ProcessorParams contains the parameters for creating a new processor.
type ProcessorParams struct {
	Logger      logger.Interface
	Service     PageService
	Validator   content.JobValidator
	Storage     types.Interface
	IndexName   string
	PageChannel chan *models.Page
}
