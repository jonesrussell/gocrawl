// Package config provides configuration management for the application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/types"
)

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *app.Config
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *log.Config
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *elasticsearch.Config
	// GetServerConfig returns the server configuration.
	GetServerConfig() *server.Config
	// GetSources returns the list of sources.
	GetSources() []types.Source
	// GetCommand returns the current command.
	GetCommand() string
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *crawler.Config
	// GetPriorityConfig returns the priority configuration.
	GetPriorityConfig() *priority.Config
}
