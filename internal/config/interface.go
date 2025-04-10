// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	logconfig "github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/config/types"
)

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *app.Config
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *logconfig.Config
	// GetServerConfig returns the server configuration.
	GetServerConfig() *server.Config
	// GetSources returns the list of sources.
	GetSources() []types.Source
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *crawler.Config
	// GetPriorityConfig returns the priority configuration.
	GetPriorityConfig() *priority.Config
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *elasticsearch.Config
	// GetStorageConfig returns the storage configuration.
	GetStorageConfig() *storage.Config
	// GetCommand returns the current command.
	GetCommand() string
	// Validate validates the configuration based on the current command.
	Validate() error
}
