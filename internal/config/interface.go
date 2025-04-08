// Package config provides configuration management for the application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
)

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *app.Config
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *log.Config
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetServerConfig returns the server configuration.
	GetServerConfig() *server.Config
	// GetSources returns the list of sources.
	GetSources() []Source
	// GetCommand returns the current command.
	GetCommand() string
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *CrawlerConfig
	// GetPriorityConfig returns the priority configuration.
	GetPriorityConfig() *priority.Config
}
