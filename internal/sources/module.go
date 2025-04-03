// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

const (
	// DefaultMaxDepth is the default maximum depth for crawling.
	DefaultMaxDepth = 2

	// DefaultRateLimit is the default rate limit for sources
	DefaultRateLimit = 5 * time.Second
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	fx.Provide(
		fx.Annotate(
			NewSourcesFromConfig,
			fx.ParamTags(`name:"config"`, ""),
		),
	),
)

// ModuleParams defines the parameters for creating a new Sources instance.
type ModuleParams struct {
	fx.In

	Config config.Interface `name:"config"`
	Logger logger.Interface
}

// Result defines the sources module's output.
type Result struct {
	fx.Out

	Sources Interface
}

// ProvideSources creates a new Sources instance.
func ProvideSources(params ModuleParams) Result {
	return Result{
		Sources: NewSourcesFromConfig(params.Config, params.Logger),
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
func NewSources(cfg *Config, logger logger.Interface) *Sources {
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
