// Package sources implements the command-line interface for managing content sources.
package sources

import (
	s "github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the sources command dependencies
var Module = fx.Module("sourcesCmd",
	s.Module,
)
