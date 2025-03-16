package config

import "time"

// MockConfig is a mock implementation of the Interface for testing.
type MockConfig struct {
	sources       []Source
	crawlerConfig *CrawlerConfig
	logConfig     *LogConfig
	appConfig     *AppConfig
	esConfig      *ElasticsearchConfig
	serverConfig  *ServerConfig
}

// NewMockConfig creates a new mock config with default values.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		sources: []Source{},
		crawlerConfig: &CrawlerConfig{
			MaxDepth:    3,
			Parallelism: 2,
			RateLimit:   time.Second,
			RandomDelay: time.Second,
		},
		logConfig: &LogConfig{
			Level: "info",
			Debug: false,
		},
		appConfig: &AppConfig{
			Environment: "test",
		},
		esConfig: &ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test_index",
		},
		serverConfig: &ServerConfig{
			Address:      ":8080",
			ReadTimeout:  time.Second * 15,
			WriteTimeout: time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
	}
}

// WithSources sets the sources for the mock config.
func (m *MockConfig) WithSources(sources []Source) *MockConfig {
	m.sources = sources
	return m
}

// WithCrawlerConfig sets the crawler config for the mock config.
func (m *MockConfig) WithCrawlerConfig(cfg *CrawlerConfig) *MockConfig {
	m.crawlerConfig = cfg
	return m
}

// WithLogConfig sets the log config for the mock config.
func (m *MockConfig) WithLogConfig(cfg *LogConfig) *MockConfig {
	m.logConfig = cfg
	return m
}

// WithAppConfig sets the app config for the mock config.
func (m *MockConfig) WithAppConfig(cfg *AppConfig) *MockConfig {
	m.appConfig = cfg
	return m
}

// WithElasticsearchConfig sets the elasticsearch config for the mock config.
func (m *MockConfig) WithElasticsearchConfig(cfg *ElasticsearchConfig) *MockConfig {
	m.esConfig = cfg
	return m
}

// WithServerConfig sets the server config for the mock config.
func (m *MockConfig) WithServerConfig(cfg *ServerConfig) *MockConfig {
	m.serverConfig = cfg
	return m
}

// GetSources implements Interface.
func (m *MockConfig) GetSources() []Source {
	return m.sources
}

// GetCrawlerConfig implements Interface.
func (m *MockConfig) GetCrawlerConfig() *CrawlerConfig {
	return m.crawlerConfig
}

// GetLogConfig implements Interface.
func (m *MockConfig) GetLogConfig() *LogConfig {
	return m.logConfig
}

// GetAppConfig implements Interface.
func (m *MockConfig) GetAppConfig() *AppConfig {
	return m.appConfig
}

// GetElasticsearchConfig implements Interface.
func (m *MockConfig) GetElasticsearchConfig() *ElasticsearchConfig {
	return m.esConfig
}

// GetServerConfig implements Interface.
func (m *MockConfig) GetServerConfig() *ServerConfig {
	return m.serverConfig
}

// Ensure MockConfig implements Interface
var _ Interface = (*MockConfig)(nil)
