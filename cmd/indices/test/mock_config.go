package test

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/stretchr/testify/mock"
)

// MockConfig is a mock implementation of config.Interface for testing.
type MockConfig struct {
	mock.Mock
}

// GetAppConfig returns the application configuration.
func (m *MockConfig) GetAppConfig() *app.Config {
	args := m.Called()
	return args.Get(0).(*app.Config)
}

// GetLogConfig returns the logging configuration.
func (m *MockConfig) GetLogConfig() *log.Config {
	args := m.Called()
	return args.Get(0).(*log.Config)
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	return args.Get(0).(*config.ElasticsearchConfig)
}

// GetServerConfig returns the server configuration.
func (m *MockConfig) GetServerConfig() *server.Config {
	args := m.Called()
	return args.Get(0).(*server.Config)
}

// GetSources returns the list of sources.
func (m *MockConfig) GetSources() []config.Source {
	args := m.Called()
	return args.Get(0).([]config.Source)
}

// GetCommand returns the current command.
func (m *MockConfig) GetCommand() string {
	args := m.Called()
	return args.String(0)
}

// GetCrawlerConfig returns the crawler configuration.
func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	args := m.Called()
	return args.Get(0).(*config.CrawlerConfig)
}

// GetPriorityConfig returns the priority configuration.
func (m *MockConfig) GetPriorityConfig() *priority.Config {
	args := m.Called()
	return args.Get(0).(*priority.Config)
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
