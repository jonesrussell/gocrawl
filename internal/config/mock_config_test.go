package config_test

import (
	"github.com/jonesrussell/gocrawl/internal/config"
)

// MockConfig is a mock implementation of config.Interface for testing.
type MockConfig struct {
	// AppConfig is the application configuration
	AppConfig *config.AppConfig
	// LogConfig is the logging configuration
	LogConfig *config.LogConfig
	// ElasticsearchConfig is the Elasticsearch configuration
	ElasticsearchConfig *config.ElasticsearchConfig
	// ServerConfig is the server configuration
	ServerConfig *config.ServerConfig
	// Sources is the list of sources
	Sources []config.Source
	// Command is the current command
	Command string
	// CrawlerConfig is the crawler configuration
	CrawlerConfig *config.CrawlerConfig
	// PriorityConfig is the priority configuration
	PriorityConfig *config.PriorityConfig
}

// WithSources sets the sources for the mock config.
func (m *MockConfig) WithSources(sources []config.Source) *MockConfig {
	m.Sources = sources
	return m
}

// WithCrawlerConfig sets the crawler config for the mock config.
func (m *MockConfig) WithCrawlerConfig(cfg *config.CrawlerConfig) *MockConfig {
	m.CrawlerConfig = cfg
	return m
}

// WithLogConfig sets the log config for the mock config.
func (m *MockConfig) WithLogConfig(cfg *config.LogConfig) *MockConfig {
	m.LogConfig = cfg
	return m
}

// WithAppConfig sets the app config for the mock config.
func (m *MockConfig) WithAppConfig(cfg *config.AppConfig) *MockConfig {
	m.AppConfig = cfg
	return m
}

// WithElasticsearchConfig sets the elasticsearch config for the mock config.
func (m *MockConfig) WithElasticsearchConfig(cfg *config.ElasticsearchConfig) *MockConfig {
	m.ElasticsearchConfig = cfg
	return m
}

// WithServerConfig sets the server config for the mock config.
func (m *MockConfig) WithServerConfig(cfg *config.ServerConfig) *MockConfig {
	m.ServerConfig = cfg
	return m
}

// GetSources returns the configured sources.
func (m *MockConfig) GetSources() []config.Source {
	return m.Sources
}

// GetCrawlerConfig returns the crawler configuration.
func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return m.CrawlerConfig
}

// GetLogConfig returns the logging configuration.
func (m *MockConfig) GetLogConfig() *config.LogConfig {
	return m.LogConfig
}

// GetAppConfig returns the application configuration.
func (m *MockConfig) GetAppConfig() *config.AppConfig {
	return m.AppConfig
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return m.ElasticsearchConfig
}

// GetServerConfig returns the server configuration.
func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	return m.ServerConfig
}

// GetCommand returns the command being run.
func (m *MockConfig) GetCommand() string {
	return m.Command
}

// GetPriorityConfig returns the priority configuration
func (m *MockConfig) GetPriorityConfig() *config.PriorityConfig {
	return &config.PriorityConfig{
		Default: 1,
		Rules:   []config.PriorityRule{},
	}
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
