// Package indices implements the command-line interface for managing Elasticsearch indices.
package indices

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// DeleteDeps holds the dependencies for the delete command
type DeleteDeps struct {
	Storage common.Storage
	Sources sources.Interface
	Logger  common.Logger
}

// Module provides the indices command dependencies
var Module = fx.Module("indices",
	// Core dependencies
	config.Module,
	sources.Module,
	storage.Module,
	logger.Module,

	// Additional modules
	api.Module,
)
