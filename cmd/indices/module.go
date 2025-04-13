// Package indices provides commands for managing Elasticsearch indices.
package indices

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the indices module for dependency injection.
var Module = fx.Module("indices",
	// Core modules
	config.Module,
	logger.Module,

	// Provide the context
	fx.Provide(context.Background),

	// Provide the indices components
	fx.Provide(
		NewCreator,
		NewLister,
		NewTableRenderer,
		NewDeleter,
	),
)
