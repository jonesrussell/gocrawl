// Package sources implements the command-line interface for managing content sources.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the sources command dependencies
var Module = fx.Module("sourcesCmd",
	// Core dependencies
	config.Module,
	logger.Module,
)
