// Package config provides configuration management for the GoCrawl application.
package config

import (
	"github.com/stretchr/testify/mock"
)

// MockConfig is a mock implementation of the Interface for testing.
type MockConfig struct {
	mock.Mock
}

// NewMockConfig creates a new MockConfig with default values.
func NewMockConfig() *MockConfig {
	m := &MockConfig{}

	// Set up default values
	m.On("GetCrawlerConfig").Return(&CrawlerConfig{
		MaxDepth:    2,
		Parallelism: 1,
	})

	m.On("GetElasticsearchConfig").Return(&ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
	})

	m.On("GetLogConfig").Return(&LogConfig{
		Level: "info",
		Debug: false,
	})

	m.On("GetAppConfig").Return(&AppConfig{
		Environment: "test",
	})

	m.On("GetServerConfig").Return(&ServerConfig{
		Address: ":8080",
	})

	// Default empty sources
	m.On("GetSources").Return([]Source{})

	return m
}

// GetCrawlerConfig implements Interface.
func (m *MockConfig) GetCrawlerConfig() *CrawlerConfig {
	args := m.Called()
	return args.Get(0).(*CrawlerConfig)
}

// GetElasticsearchConfig implements Interface.
func (m *MockConfig) GetElasticsearchConfig() *ElasticsearchConfig {
	args := m.Called()
	return args.Get(0).(*ElasticsearchConfig)
}

// GetLogConfig implements Interface.
func (m *MockConfig) GetLogConfig() *LogConfig {
	args := m.Called()
	return args.Get(0).(*LogConfig)
}

// GetAppConfig implements Interface.
func (m *MockConfig) GetAppConfig() *AppConfig {
	args := m.Called()
	return args.Get(0).(*AppConfig)
}

// GetSources implements Interface.
func (m *MockConfig) GetSources() []Source {
	args := m.Called()
	return args.Get(0).([]Source)
}

// GetServerConfig implements Interface.
func (m *MockConfig) GetServerConfig() *ServerConfig {
	args := m.Called()
	return args.Get(0).(*ServerConfig)
}

// WithSources sets up the mock to return the specified sources.
func (m *MockConfig) WithSources(sources []Source) *MockConfig {
	m.On("GetSources").Return(sources)
	return m
}

// WithCrawlerConfig sets up the mock to return the specified crawler config.
func (m *MockConfig) WithCrawlerConfig(cfg *CrawlerConfig) *MockConfig {
	m.On("GetCrawlerConfig").Return(cfg)
	return m
}

// WithElasticsearchConfig sets up the mock to return the specified Elasticsearch config.
func (m *MockConfig) WithElasticsearchConfig(cfg *ElasticsearchConfig) *MockConfig {
	m.On("GetElasticsearchConfig").Return(cfg)
	return m
}

// WithLogConfig sets up the mock to return the specified log config.
func (m *MockConfig) WithLogConfig(cfg *LogConfig) *MockConfig {
	m.On("GetLogConfig").Return(cfg)
	return m
}

// WithAppConfig sets up the mock to return the specified app config.
func (m *MockConfig) WithAppConfig(cfg *AppConfig) *MockConfig {
	m.On("GetAppConfig").Return(cfg)
	return m
}

// WithServerConfig sets up the mock to return the specified server config.
func (m *MockConfig) WithServerConfig(cfg *ServerConfig) *MockConfig {
	m.On("GetServerConfig").Return(cfg)
	return m
}
