// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
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
	GetSources() []Config
}

// Module provides the sources module's dependencies.
var Module = fx.Module("sources",
	fx.Provide(
		fx.Annotate(
			provideSources,
			fx.ResultTags(`name:"sourceManager"`),
		),
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

// provideSources creates a new sources instance.
func provideSources(p Params) Interface {
	var configs []Config
	sources := p.Config.GetSources()
	for _, src := range sources {
		configs = append(configs, Config{
			Name:         src.Name,
			URL:          src.URL,
			RateLimit:    src.RateLimit,
			MaxDepth:     src.MaxDepth,
			ArticleIndex: src.ArticleIndex,
			Index:        src.Index,
			Time:         src.Time,
			Selectors:    NewSelectorConfigFromSource(src),
			Metadata:     src.Metadata,
		})
	}
	if len(configs) == 0 {
		configs = append(configs, *NewConfig())
	}
	return &Sources{
		sources: configs,
	}
}

// NewConfig creates a new source configuration.
func NewConfig() *Config {
	return &Config{
		Name:     "default",
		URL:      "http://localhost",
		MaxDepth: DefaultMaxDepth,
	}
}

// NewSources creates a new sources instance.
func NewSources(cfg *Config) Interface {
	return &Sources{
		sources: []Config{*cfg},
	}
}
