// Package indices implements the command-line interface for managing Elasticsearch indices.
package indices

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// DeleteDeps holds the dependencies for the delete command
type DeleteDeps struct {
	Storage types.Interface
	Sources sources.Interface
	Logger  logger.Interface
}

// Module provides the indices command dependencies
var Module = fx.Module("indices",
	// Core dependencies
	config.Module,
	logger.Module,
	sources.Module,
	storage.Module,
)
