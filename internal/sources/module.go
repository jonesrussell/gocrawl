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

	// DefaultRateLimit is the default rate limit for sources
	DefaultRateLimit = 5 * time.Second
)

// Module provides the sources module's dependencies.
var Module = fx.Module("sources",
	fx.Provide(
		ProvideSources,
	),
)

// ModuleParams holds the parameters for creating a sources module.
type ModuleParams struct {
	fx.In

	Config config.Interface
	Logger Logger
}

// ProvideSources creates a new sources instance.
func ProvideSources(p ModuleParams) Interface {
	var configs []Config
	sources := p.Config.GetSources()
	if len(sources) == 0 {
		// If no sources found, use a default config
		defaultConfig := NewConfig()
		defaultConfig.Name = "default"
		defaultConfig.URL = "http://localhost"
		defaultConfig.MaxDepth = DefaultMaxDepth
		defaultConfig.RateLimit = DefaultRateLimit
		defaultConfig.ArticleIndex = "articles"
		defaultConfig.Index = "content"
		configs = append(configs, *defaultConfig)
	} else {
		for _, src := range sources {
			config := Config{
				Name:         src.Name,
				URL:          src.URL,
				RateLimit:    src.RateLimit,
				MaxDepth:     src.MaxDepth,
				Time:         src.Time,
				ArticleIndex: src.ArticleIndex,
				Index:        src.Index,
			}
			// Set default index names if empty
			if config.ArticleIndex == "" {
				config.ArticleIndex = "articles"
			}
			if config.Index == "" {
				config.Index = "content"
			}
			configs = append(configs, config)
		}
	}
	return &Sources{
		sources: configs,
		logger:  p.Logger,
		metrics: Metrics{
			SourceCount: int64(len(configs)),
		},
	}
}

// NewConfig creates a new source configuration.
func NewConfig() *Config {
	return &Config{
		Name:         "default",
		URL:          "http://localhost",
		MaxDepth:     DefaultMaxDepth,
		RateLimit:    DefaultRateLimit,
		ArticleIndex: "articles",
		Index:        "content",
	}
}

// NewSources creates a new sources instance.
func NewSources(cfg *Config, logger Logger) *Sources {
	return &Sources{
		sources: []Config{*cfg},
		logger:  logger,
		metrics: Metrics{
			SourceCount: 1,
		},
	}
}

// DefaultConfig returns the default source configuration.
func DefaultConfig() *Config {
	defaultConfig := &Config{
		RateLimit:    DefaultRateLimit,
		ArticleIndex: "articles",
		Index:        "content",
	}
	return defaultConfig
}
