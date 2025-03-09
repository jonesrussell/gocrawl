// Package config provides configuration management for the GoCrawl application.
package config

// Interface defines the configuration interface.
type Interface interface {
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *CrawlerConfig

	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig

	// GetLogConfig returns the logging configuration.
	GetLogConfig() *LogConfig

	// GetAppConfig returns the application configuration.
	GetAppConfig() *AppConfig

	// GetSources returns the list of sources.
	GetSources() []Source

	// GetServerConfig returns the server configuration.
	GetServerConfig() *ServerConfig
}
