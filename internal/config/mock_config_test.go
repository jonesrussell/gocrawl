package config_test

import (
	"github.com/jonesrussell/gocrawl/internal/config"
)

// MockConfig is a mock implementation of config.Interface for testing.
type MockConfig struct {
	sources             []config.Source
	crawlerConfig       *config.CrawlerConfig
	logConfig           *config.LogConfig
	appConfig           *config.AppConfig
	elasticsearchConfig *config.ElasticsearchConfig
	serverConfig        *config.ServerConfig
}

// WithSources sets the sources for the mock config.
func (m *MockConfig) WithSources(sources []config.Source) *MockConfig {
	m.sources = sources
	return m
}

// WithCrawlerConfig sets the crawler config for the mock config.
func (m *MockConfig) WithCrawlerConfig(cfg *config.CrawlerConfig) *MockConfig {
	m.crawlerConfig = cfg
	return m
}

// WithLogConfig sets the log config for the mock config.
func (m *MockConfig) WithLogConfig(cfg *config.LogConfig) *MockConfig {
	m.logConfig = cfg
	return m
}

// WithAppConfig sets the app config for the mock config.
func (m *MockConfig) WithAppConfig(cfg *config.AppConfig) *MockConfig {
	m.appConfig = cfg
	return m
}

// WithElasticsearchConfig sets the elasticsearch config for the mock config.
func (m *MockConfig) WithElasticsearchConfig(cfg *config.ElasticsearchConfig) *MockConfig {
	m.elasticsearchConfig = cfg
	return m
}

// WithServerConfig sets the server config for the mock config.
func (m *MockConfig) WithServerConfig(cfg *config.ServerConfig) *MockConfig {
	m.serverConfig = cfg
	return m
}

// GetSources returns the configured sources.
func (m *MockConfig) GetSources() []config.Source {
	return m.sources
}

// GetCrawlerConfig returns the crawler configuration.
func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return m.crawlerConfig
}

// GetLogConfig returns the logging configuration.
func (m *MockConfig) GetLogConfig() *config.LogConfig {
	return m.logConfig
}

// GetAppConfig returns the application configuration.
func (m *MockConfig) GetAppConfig() *config.AppConfig {
	return m.appConfig
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return m.elasticsearchConfig
}

// GetServerConfig returns the server configuration.
func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	return m.serverConfig
}

// GetCommand returns the command being run.
func (m *MockConfig) GetCommand() string {
	return "test"
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
