// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	// Include core modules
	sources.Module,
	// Provide the source components
	fx.Provide(
		NewTableRenderer,
		NewLister,
	),
)
