// Package index provides commands for managing Elasticsearch index.
package index

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the index module for dependency injection.
var Module = fx.Module("index",
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	sources.Module,

	// Provide the context
	fx.Provide(context.Background),

	// Provide the index components
	fx.Provide(
		NewCreator,
		NewLister,
		NewTableRenderer,
		NewDeleter,
	),
)
