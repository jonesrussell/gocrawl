// Package sources provides the sources module for dependency injection.
package sources

import (
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"go.uber.org/fx"
)

// SourcesParams contains dependencies for creating sources
type SourcesParams struct {
	fx.In

	Logger logger.Interface
	Config config.Interface
}

// SourcesResult contains the sources and its components
type SourcesResult struct {
	fx.Out

	Sources Interface
}

// NewSourcesProvider creates a new Sources instance
func NewSourcesProvider(p SourcesParams) (SourcesResult, error) {
	sources, err := LoadSources(p.Config, p.Logger)
	if err != nil {
		return SourcesResult{}, err
	}

	return SourcesResult{
		Sources: sources,
	}, nil
}

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	fx.Provide(
		NewSourcesProvider,
	),
)

// ModuleParams defines the parameters for creating a new Sources instance.
type ModuleParams struct {
	fx.In

	Deps cmdcommon.CommandDeps
}

// Result defines the output of the sources module.
type Result struct {
	fx.Out

	Sources Interface
}

// ProvideSources creates a new Sources instance from the given configuration.
func ProvideSources(params ModuleParams) (Interface, error) {
	sources, err := LoadSources(params.Deps.Config, params.Deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	return sources, nil
}

// NewSources creates a new sources instance.
func NewSources(cfg *Config, deps cmdcommon.CommandDeps) *Sources {
	return &Sources{
		sources: []Config{*cfg},
		logger:  deps.Logger,
		metrics: sourceutils.NewSourcesMetrics(),
	}
}
