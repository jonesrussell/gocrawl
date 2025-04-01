// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

const (
	// DefaultMaxDepth is the default maximum depth for crawling.
	DefaultMaxDepth = 2
)

// Interface defines the interface for managing web content sources.
type Interface interface {
	// FindByName finds a source by its name.
	FindByName(name string) (*Config, error)

	// Validate checks if a source configuration is valid.
	Validate(source *Config) error

	// GetSources returns all available sources.
	GetSources() ([]Config, error)
}

// Module provides the sources module's dependencies.
var Module = fx.Module("sources",
	fx.Provide(
		provideSources,
	),
)

// Params defines the required dependencies for the sources module.
type Params struct {
	fx.In

	Config config.Interface
}

// provideSources creates a new sources instance.
func provideSources(p Params) Interface {
	var configs []Config
	sources := p.Config.GetSources()
	if len(sources) == 0 {
		// If no sources found, use a default config
		defaultConfig := NewConfig()
		defaultConfig.Name = "default"
		defaultConfig.URL = "http://localhost"
		defaultConfig.MaxDepth = DefaultMaxDepth
		defaultConfig.RateLimit = 5 * time.Second
		configs = append(configs, *defaultConfig)
	} else {
		for _, src := range sources {
			configs = append(configs, Config{
				Name:      src.Name,
				URL:       src.URL,
				RateLimit: src.RateLimit,
				MaxDepth:  src.MaxDepth,
				Time:      src.Time,
			})
		}
	}
	return &Sources{
		sources: configs,
	}
}

// NewConfig creates a new source configuration.
func NewConfig() *Config {
	return &Config{
		Name:      "default",
		URL:       "http://localhost",
		MaxDepth:  DefaultMaxDepth,
		RateLimit: 5 * time.Second,
	}
}

// NewSources creates a new sources instance.
func NewSources(cfg *Config) Interface {
	return &Sources{
		sources: []Config{*cfg},
	}
}
