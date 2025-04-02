// Package indices implements the command-line interface for managing Elasticsearch indices.
package indices

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// DeleteDeps holds the dependencies for the delete command
type DeleteDeps struct {
	fx.In

	Storage storagetypes.Interface
	Sources sources.Interface `name:"sources"`
	Logger  common.Logger
}

// CreateParams holds the parameters for creating an index
type CreateParams struct {
	fx.In

	Context context.Context
	Logger  types.Logger
	Storage storagetypes.Interface
	Config  config.Interface
}

// Module provides the indices command dependencies
var Module = fx.Module("indices",
	// Core dependencies
	config.Module,
	logger.Module,
	sources.Module,
	storage.Module,
)
