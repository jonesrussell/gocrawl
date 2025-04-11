// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"go.uber.org/fx"
)

// Module provides the configuration module for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		// Provide both concrete type and interface
		fx.Annotate(
			LoadConfig,
			fx.As(new(Interface)),
		),
		// Provide crawler config
		func(cfg Interface) *crawler.Config {
			return cfg.GetCrawlerConfig()
		},
	),
)
