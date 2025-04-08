package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
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
			BaseURL:     "http://test.com",
			MaxDepth:    defaultMaxDepth,
			RateLimit:   time.Second,
			RandomDelay: time.Second,
			IndexName:   "test_index",
			SourceFile:  "", // Empty to prevent loading from file
		}
	}
	return args.Get(0).(*config.CrawlerConfig)
}

// GetLogConfig implements config.Interface.
func (m *MockConfig) GetLogConfig() *log.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return &log.Config{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}
	return args.Get(0).(*log.Config)
}

// GetAppConfig implements config.Interface.
func (m *MockConfig) GetAppConfig() *app.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return &app.Config{
			Environment: "test",
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       false,
		}
	}
	return args.Get(0).(*app.Config)
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
func (m *MockConfig) GetServerConfig() *server.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return &server.Config{
			SecurityEnabled: false,
			APIKey:          "",
		}
	}
	return args.Get(0).(*server.Config)
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
func (m *MockConfig) GetPriorityConfig() *priority.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return &priority.Config{
			Default: 1,
			Rules:   []priority.Rule{},
		}
	}
	return args.Get(0).(*priority.Config)
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
