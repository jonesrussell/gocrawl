// Package content provides functionality for processing and managing general web content.
// It includes services for content extraction, processing, and storage, with support
// for different content types and formats. This package is designed to handle
// non-article content that may be encountered during web crawling.
package content

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProcessorParams contains the dependencies required to create a content processor.
type ProcessorParams struct {
	fx.In

	Logger    common.Logger
	Service   Interface
	Storage   types.Interface
	IndexName string `name:"contentIndex"`
}

// Module provides the content module's dependencies.
var Module = fx.Module("content",
	fx.Provide(
		// Provide content processor with all dependencies
		fx.Annotate(
			func(p ProcessorParams) common.Processor {
				// Create processor
				processor := NewProcessor(
					p.Service,
					p.Storage,
					p.Logger,
					p.IndexName,
				)
				p.Logger.Debug("Created content processor", "type", fmt.Sprintf("%T", processor))
				return processor
			},
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// Params defines the parameters required for creating a content service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
// - Config: For application configuration
// - Storage: For content persistence
// - IndexName: The Elasticsearch index name for content
type Params struct {
	fx.In

	Logger    common.Logger
	Config    *config.Config
	Storage   types.Interface
	IndexName string
}
