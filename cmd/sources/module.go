// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the sources command functionality.
var Module = fx.Module("sources",
	// Include core modules
	sources.Module,
)
