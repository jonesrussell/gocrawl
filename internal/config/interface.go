// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/logging"
	"github.com/jonesrussell/gocrawl/internal/config/server"
)

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *app.Config
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *logging.Config
	// GetServerConfig returns the server configuration.
	GetServerConfig() *server.Config
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *crawler.Config
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *elasticsearch.Config
	// GetCommand returns the current command.
	GetCommand() string
	// GetConfigFile returns the path to the configuration file.
	GetConfigFile() string
	// Validate validates the configuration based on the current command.
	Validate() error
}
