// Package page provides functionality for processing and managing web pages.
package page

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ServiceParams contains the parameters for creating a new content service.
type ServiceParams struct {
	Logger    logger.Interface
	Storage   types.Interface
	IndexName string
}

// ProcessorParams contains the parameters for creating a new processor.
type ProcessorParams struct {
	Logger      logger.Interface
	Service     Interface
	Validator   content.JobValidator
	Storage     types.Interface
	IndexName   string
	PageChannel chan *domain.Page
}
