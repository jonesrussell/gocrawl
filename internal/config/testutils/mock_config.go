package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/mock"
)

const (
	defaultMaxDepth     = 3
	defaultParallelism  = 2
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

// MockConfig is a mock implementation of the config.Interface for testing.
type MockConfig struct {
	mock.Mock
}

// GetSources implements config.Interface.
func (m *MockConfig) GetSources() []config.Source {
	args := m.Called()
	if args.Get(0) == nil {
		return []config.Source{}
	}
	return args.Get(0).([]config.Source)
}

// GetCrawlerConfig implements config.Interface.
func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.CrawlerConfig{
			BaseURL:          "http://test.com",
			MaxDepth:         defaultMaxDepth,
			RateLimit:        time.Second,
			RandomDelay:      time.Second,
			IndexName:        "test_index",
			ContentIndexName: "test_content",
			SourceFile:       "", // Empty to prevent loading from file
		}
	}
	return args.Get(0).(*config.CrawlerConfig)
}

// GetLogConfig implements config.Interface.
func (m *MockConfig) GetLogConfig() *config.LogConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.LogConfig{
			Level: "info",
			Debug: false,
		}
	}
	return args.Get(0).(*config.LogConfig)
}

// GetAppConfig implements config.Interface.
func (m *MockConfig) GetAppConfig() *config.AppConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.AppConfig{
			Environment: "test",
		}
	}
	return args.Get(0).(*config.AppConfig)
}

// GetElasticsearchConfig implements config.Interface.
func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test_index",
		}
	}
	return args.Get(0).(*config.ElasticsearchConfig)
}

// GetServerConfig implements config.Interface.
func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.ServerConfig{
			Address:      ":8080",
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			IdleTimeout:  defaultIdleTimeout,
		}
	}
	return args.Get(0).(*config.ServerConfig)
}

// GetCommand implements config.Interface.
func (m *MockConfig) GetCommand() string {
	args := m.Called()
	if args.Get(0) == nil {
		return "test"
	}
	return args.String(0)
}

// GetPriorityConfig implements config.Interface.
func (m *MockConfig) GetPriorityConfig() *config.PriorityConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return &config.PriorityConfig{
			Default: 1,
			Rules:   []config.PriorityRule{},
		}
	}
	return args.Get(0).(*config.PriorityConfig)
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
