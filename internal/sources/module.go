// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

const (
	// DefaultMaxDepth is the default maximum depth for crawling.
	DefaultMaxDepth = 2

	// DefaultRateLimit is the default rate limit for sources
	DefaultRateLimit = 5 * time.Second
)

// Module provides the sources package's dependencies.
var Module = fx.Module("sources",
	fx.Provide(
		// Provide sources from config
		func(cfg config.Interface, logger Logger) Interface {
			return NewSourcesFromConfig(cfg, logger)
		},
		// Provide logger
		func(logger types.Logger) Logger {
			return logger
		},
	),
)

// ModuleParams defines the parameters for creating a new Sources instance.
type ModuleParams struct {
	fx.In

	Config config.Interface
	Logger types.Logger
}

// Result defines the sources module's output.
type Result struct {
	fx.Out

	Sources Interface
}

// ProvideSources creates a new Sources instance.
func ProvideSources(p ModuleParams) (Result, error) {
	// Get sources from config
	configSources := p.Config.GetSources()

	// Convert config sources to internal sources
	var sources []Config
	for _, src := range configSources {
		// Set default index names if empty
		if src.ArticleIndex == "" {
			src.ArticleIndex = "articles"
		}
		if src.Index == "" {
			src.Index = "content"
		}

		sources = append(sources, convertSourceConfig(src))
	}

	// Create sources instance
	s := &Sources{
		sources: sources,
		logger:  p.Logger,
		metrics: Metrics{
			SourceCount: int64(len(sources)),
		},
	}

	return Result{
		Sources: s,
	}, nil
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
