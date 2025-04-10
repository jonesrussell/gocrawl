package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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
func (m *MockConfig) GetSources() []types.Source {
	args := m.Called()
	sources, _ := args.Get(0).([]types.Source)
	return sources
}

// GetCrawlerConfig implements config.Interface.
func (m *MockConfig) GetCrawlerConfig() *crawler.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*crawler.Config)
	return cfg
}

// GetLogConfig implements config.Interface.
func (m *MockConfig) GetLogConfig() *log.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*log.Config)
	return cfg
}

// GetAppConfig implements config.Interface.
func (m *MockConfig) GetAppConfig() *app.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*app.Config)
	return cfg
}

// GetElasticsearchConfig implements config.Interface.
func (m *MockConfig) GetElasticsearchConfig() *elasticsearch.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*elasticsearch.Config)
	return cfg
}

// GetServerConfig implements config.Interface.
func (m *MockConfig) GetServerConfig() *server.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*server.Config)
	return cfg
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
	cfg, _ := args.Get(0).(*priority.Config)
	return cfg
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)
