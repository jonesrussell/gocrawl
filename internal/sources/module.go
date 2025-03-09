// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

// Module provides the sources module and its dependencies.
var Module = fx.Module("sources",
	fx.Provide(
		provideSourceConfig,
	),
)

// Params defines the required dependencies for the sources module.
type Params struct {
	fx.In

	Config config.Interface
}

// Result contains the components provided by the sources module.
type Result struct {
	fx.Out

	Sources *Sources
}

// provideSourceConfig creates a new Sources instance from configuration.
func provideSourceConfig(p Params) (Result, error) {
	sources, err := LoadFromFile(p.Config.GetCrawlerConfig().SourceFile)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Sources: sources,
	}, nil
}
