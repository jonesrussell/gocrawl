// Package content provides functionality for processing and managing general web content.
// It includes services for content extraction, processing, and storage, with support
// for different content types and formats. This package is designed to handle
// non-article content that may be encountered during web crawling.
package content

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProcessorParams contains the dependencies required to create a content processor.
type ProcessorParams struct {
	fx.In

	Logger    logger.Interface
	Service   Interface
	Storage   types.Interface
	IndexName string `name:"contentIndex"`
}

// Params defines the parameters required for creating a content service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
type Params struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"contentIndexName"`
}

// Module provides the content module's dependencies.
var Module = fx.Module("content",
	fx.Provide(
		// Provide the content service
		func(p Params) (Interface, error) {
			service := NewService(p.Logger)
			p.Logger.Debug("Created content service", "type", fmt.Sprintf("%T", service))
			return service, nil
		},
		// Provide content processor for the group
		fx.Annotate(
			func(p ProcessorParams) common.Processor {
				// Create processor
				processor := NewContentProcessor(p)
				p.Logger.Debug("Created content processor", "type", fmt.Sprintf("%T", processor))
				return processor
			},
			fx.ResultTags(`group:"processors"`),
		),
		// Provide content processor with name
		fx.Annotate(
			func(p ProcessorParams) common.Processor {
				// Create processor
				processor := NewContentProcessor(p)
				p.Logger.Debug("Created content processor", "type", fmt.Sprintf("%T", processor))
				return processor
			},
			fx.ResultTags(`name:"contentProcessor"`),
		),
	),
)
