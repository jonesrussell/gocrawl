// Package config provides configuration management for the GoCrawl application.
package config

// Interface defines the configuration interface that must be implemented
// by any configuration provider.
type Interface interface {
	// GetAppConfig returns the application configuration
	GetAppConfig() *AppConfig
	// GetCrawlerConfig returns the crawler configuration
	GetCrawlerConfig() *CrawlerConfig
	// GetElasticsearchConfig returns the Elasticsearch configuration
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetLogConfig returns the logging configuration
	GetLogConfig() *LogConfig
	// GetServerConfig returns the server configuration
	GetServerConfig() *ServerConfig
	// GetSources returns the list of sources to crawl
	GetSources() []Source
	// GetCommand returns the command being run
	GetCommand() string
}
