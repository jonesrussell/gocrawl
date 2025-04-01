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
	"github.com/jonesrussell/gocrawl/pkg/collector"
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
			func(
				logger common.Logger,
				storage types.Interface,
				params struct {
					fx.In
					IndexName string `name:"contentIndex"`
				},
			) collector.Processor {
				// Create service
				service := NewService(logger)
				logger.Debug("Created content service", "type", fmt.Sprintf("%T", service))

				// Create processor
				processor := &ContentProcessor{
					service:   service,
					storage:   storage,
					logger:    logger,
					indexName: params.IndexName,
				}
				logger.Debug("Created content processor", "type", fmt.Sprintf("%T", processor))
				return processor
			},
			fx.ResultTags(`name:"contentProcessor"`),
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
