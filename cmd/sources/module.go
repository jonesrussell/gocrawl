// Package sources implements the command-line interface for managing content sources.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"go.uber.org/fx"
)

// Module provides the sources command dependencies
var Module = fx.Module("sourcesCmd",
	// Core dependencies
	common.Module,
)
