// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	// Include required modules
	cmdcommon.Module,
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
	sources, err := LoadSources(params.Deps.Config)
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
