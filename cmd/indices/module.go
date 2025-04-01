// Package indices implements the command-line interface for managing Elasticsearch indices.
package indices

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// DeleteDeps holds the dependencies for the delete command
type DeleteDeps struct {
	fx.In

	Storage types.Interface
	Sources sources.Interface `name:"sources"`
	Logger  common.Logger
}

// Module provides the indices command dependencies
var Module = fx.Module("indices",
	// Core dependencies
	config.Module,
	logger.Module,
	sources.Module,
	storage.Module,
)
