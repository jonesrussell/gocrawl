// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

// Interface defines the interface for managing web content sources.
type Interface interface {
	// FindByName finds a source by its name.
	FindByName(name string) (*Config, error)

	// Validate checks if a source configuration is valid.
	Validate(source *Config) error

	// GetSources returns all available sources.
	GetSources() []Config
}

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

	Sources Interface `name:"sourceManager"`
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
