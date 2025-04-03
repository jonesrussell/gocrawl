// Package config provides configuration management for the application.
package config

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *AppConfig
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *LogConfig
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetServerConfig returns the server configuration.
	GetServerConfig() *ServerConfig
	// GetSources returns the list of sources.
	GetSources() []Source
	// GetCommand returns the current command.
	GetCommand() string
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *CrawlerConfig
	// GetPriorityConfig returns the priority configuration.
	GetPriorityConfig() *PriorityConfig
}
