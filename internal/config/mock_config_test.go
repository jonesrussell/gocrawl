package config_test

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
)

// MockConfig is a mock implementation of config.Interface for testing.
type MockConfig struct {
	// AppConfig is the application configuration
	AppConfig *app.Config
	// LogConfig is the logging configuration
	LogConfig *log.Config
	// ElasticsearchConfig is the Elasticsearch configuration
	ElasticsearchConfig *elasticsearch.Config
	// ServerConfig is the server configuration
	ServerConfig *server.Config
	// Sources is the list of sources
	Sources []config.Source
	// Command is the current command
	Command string
	// CrawlerConfig is the crawler configuration
	CrawlerConfig *config.CrawlerConfig
	// PriorityConfig is the priority configuration
	PriorityConfig *priority.Config
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
func (m *MockConfig) WithLogConfig(cfg *log.Config) *MockConfig {
	m.LogConfig = cfg
	return m
}

// WithAppConfig sets the app config for the mock config.
func (m *MockConfig) WithAppConfig(cfg *app.Config) *MockConfig {
	m.AppConfig = cfg
	return m
}

// WithElasticsearchConfig sets the elasticsearch config for the mock config.
func (m *MockConfig) WithElasticsearchConfig(cfg *elasticsearch.Config) *MockConfig {
	m.ElasticsearchConfig = cfg
	return m
}

// WithServerConfig sets the server config for the mock config.
func (m *MockConfig) WithServerConfig(cfg *server.Config) *MockConfig {
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
func (m *MockConfig) GetLogConfig() *log.Config {
	return m.LogConfig
}

// GetAppConfig returns the application configuration.
func (m *MockConfig) GetAppConfig() *app.Config {
	return m.AppConfig
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (m *MockConfig) GetElasticsearchConfig() *elasticsearch.Config {
	return m.ElasticsearchConfig
}

// GetServerConfig returns the server configuration.
func (m *MockConfig) GetServerConfig() *server.Config {
	return m.ServerConfig
}

// GetCommand returns the command being run.
func (m *MockConfig) GetCommand() string {
	return m.Command
}

// GetPriorityConfig returns the priority configuration
func (m *MockConfig) GetPriorityConfig() *priority.Config {
	return &priority.Config{
		DefaultPriority: 1,
		Rules:           []priority.Rule{},
	}
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
