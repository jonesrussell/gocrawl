// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// ConfigParams contains dependencies for creating configuration
type ConfigParams struct {
	fx.In

	Logger logger.Interface
}

// ConfigResult contains all configuration components
type ConfigResult struct {
	fx.Out

	CrawlerConfig       *crawler.Config
	ElasticsearchConfig *elasticsearch.Config
	ServerConfig        *server.Config
}

// ProvideConfig creates all configuration components
func ProvideConfig(p ConfigParams) ConfigResult {
	return ConfigResult{
		CrawlerConfig:       crawler.New(),
		ElasticsearchConfig: elasticsearch.NewConfig(),
		ServerConfig:        server.NewConfig(),
	}
}

// Module provides the configuration module for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		ProvideConfig,
	),
)
