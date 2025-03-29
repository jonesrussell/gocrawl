// Package content provides functionality for processing and managing general web content.
// It includes services for content extraction, processing, and storage, with support
// for different content types and formats. This package is designed to handle
// non-article content that may be encountered during web crawling.
package content

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the content service module for dependency injection.
// It provides:
// - Content service instance
// - Content processor instance
// The module uses fx.Provide to wire up dependencies and ensure proper
// initialization of content-related components.
var Module = fx.Module("content",
	fx.Provide(
		NewService,
		NewProcessor,
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
