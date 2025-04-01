// Package config provides configuration management for the GoCrawl application.
package config

// Interface defines the interface for configuration operations.
// It provides access to application configuration settings and
// supports different configuration sources.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *AppConfig
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *LogConfig
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetServerConfig returns the server configuration.
	GetServerConfig() *ServerConfig
	// GetSources returns the list of configured sources.
	GetSources() []Source
	// GetCommand returns the current command being executed.
	GetCommand() string
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *CrawlerConfig
}
