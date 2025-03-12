// Package sources implements the command-line interface for managing content sources.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the sources command dependencies
var Module = fx.Module("sourcesCmd",
	common.Module,
	sources.Module,
)
