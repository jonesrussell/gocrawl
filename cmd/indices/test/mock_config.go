package test

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/mock"
)

// MockConfig implements config.Interface for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetCommand() string {
	args := m.Called()
	return args.String(0)
}

// Unused methods that satisfy the interface
func (m *MockConfig) GetAppConfig() *config.AppConfig {
	return nil
}

func (m *MockConfig) GetLogConfig() *config.LogConfig {
	return nil
}

func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return nil
}

func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	return nil
}

func (m *MockConfig) GetSources() []config.Source {
	return nil
}

func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return nil
}
