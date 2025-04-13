// Package sources provides the sources command implementation.
package sources

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the sources command functionality.
var Module = fx.Module("sources",
	// Core modules
	config.Module,
	logger.Module,
	sources.Module,

	// Providers
	fx.Provide(
		// Provide context
		context.Background,

		// Provide command dependencies
		NewSourcesCommand,
	),
)
