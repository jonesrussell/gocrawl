package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// MockConfig is a mock implementation of the config.Interface for testing.
type MockConfig struct {
	sources       []config.Source
	crawlerConfig *config.CrawlerConfig
	logConfig     *config.LogConfig
	appConfig     *config.AppConfig
	esConfig      *config.ElasticsearchConfig
	serverConfig  *config.ServerConfig
}

// NewMockConfig creates a new mock config with default values.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		sources: []config.Source{},
		crawlerConfig: &config.CrawlerConfig{
			MaxDepth:    3,
			Parallelism: 2,
			RateLimit:   time.Second,
			RandomDelay: time.Second,
		},
		logConfig: &config.LogConfig{
			Level: "info",
			Debug: false,
		},
		appConfig: &config.AppConfig{
			Environment: "test",
		},
		esConfig: &config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test_index",
		},
		serverConfig: &config.ServerConfig{
			Address:      ":8080",
			ReadTimeout:  time.Second * 15,
			WriteTimeout: time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
	}
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
	m.esConfig = cfg
	return m
}

// WithServerConfig sets the server config for the mock config.
func (m *MockConfig) WithServerConfig(cfg *config.ServerConfig) *MockConfig {
	m.serverConfig = cfg
	return m
}

// GetSources implements config.Interface.
func (m *MockConfig) GetSources() []config.Source {
	return m.sources
}

// GetCrawlerConfig implements config.Interface.
func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return m.crawlerConfig
}

// GetLogConfig implements config.Interface.
func (m *MockConfig) GetLogConfig() *config.LogConfig {
	return m.logConfig
}

// GetAppConfig implements config.Interface.
func (m *MockConfig) GetAppConfig() *config.AppConfig {
	return m.appConfig
}

// GetElasticsearchConfig implements config.Interface.
func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return m.esConfig
}

// GetServerConfig implements config.Interface.
func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	return m.serverConfig
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
