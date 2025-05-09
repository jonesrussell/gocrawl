// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"go.uber.org/fx"
)

// Module provides the configuration module for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		// Provide the concrete config type
		LoadConfig,
		// Provide the interface
		func(cfg *Config) Interface {
			return cfg
		},
		// Provide specific configs
		func(cfg *Config) *crawler.Config {
			return cfg.GetCrawlerConfig()
		},
		func(cfg *Config) *elasticsearch.Config {
			return cfg.GetElasticsearchConfig()
		},
		func(cfg *Config) *server.Config {
			return cfg.GetServerConfig()
		},
		func(cfg *Config) *app.Config {
			return cfg.GetAppConfig()
		},
	),
)
