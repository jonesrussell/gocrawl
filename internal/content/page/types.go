// Package page provides functionality for processing and managing web pages.
package page

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

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
	Service     Interface
	JobService  common.JobService
	Storage     types.Interface
	IndexName   string
	PageChannel chan *models.Page
}
