// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
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
			ProvideSources,
			fx.As(new(Interface)),
		),
	),
)

// ModuleParams defines the parameters for creating a new Sources instance.
type ModuleParams struct {
	fx.In

	Config config.Interface
	Logger logger.Interface
}

// Result defines the output of the sources module.
type Result struct {
	fx.Out

	Sources Interface
}

// ProvideSources creates a new Sources instance from the given configuration.
func ProvideSources(params ModuleParams) (Interface, error) {
	sources, err := LoadSources(params.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	return sources, nil
}

// NewConfig creates a new source configuration.
func NewConfig() *Config {
	return &Config{
		Name:           "default",
		URL:            "http://localhost",
		AllowedDomains: []string{"localhost"},
		StartURLs:      []string{"http://localhost"},
		MaxDepth:       DefaultMaxDepth,
		RateLimit:      DefaultRateLimit,
		Rules:          types.Rules{},
	}
}

// NewSources creates a new sources instance.
func NewSources(cfg *Config, logger logger.Interface) *Sources {
	return &Sources{
		sources: []Config{*cfg},
		logger:  logger,
		metrics: sourceutils.NewSourcesMetrics(),
	}
}

// DefaultConfig returns the default source configuration.
func DefaultConfig() *Config {
	defaultConfig := &Config{
		RateLimit: DefaultRateLimit,
		Rules:     types.Rules{},
	}
	return defaultConfig
}
