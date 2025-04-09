// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/stretchr/testify/mock"
)

const (
	defaultMaxDepth    = 2
	defaultParallelism = 2
	defaultRandomDelay = 500 * time.Millisecond
	defaultRateLimit   = 100
	defaultMaxAge      = 86400
)

// MockConfig implements config.Interface for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetAppConfig() *app.Config {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if appCfg, ok := cfg.(*app.Config); ok {
			return appCfg
		}
	}
	return nil
}

func (m *MockConfig) GetLogConfig() *config.LogConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if logCfg, ok := cfg.(*config.LogConfig); ok {
			return logCfg
		}
	}
	return nil
}

func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return &config.CrawlerConfig{
		SourceFile:  "internal/api/testdata/sources.yml",
		MaxDepth:    defaultMaxDepth,
		RateLimit:   1 * time.Second,
		RandomDelay: defaultRandomDelay,
		Parallelism: defaultParallelism,
	}
}

func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if esCfg, ok := cfg.(*config.ElasticsearchConfig); ok {
			return esCfg
		}
	}
	return nil
}

func (m *MockConfig) GetServerConfig() *server.Config {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if serverCfg, ok := cfg.(*server.Config); ok {
			return serverCfg
		}
	}
	return nil
}

func (m *MockConfig) GetSources() []config.Source {
	args := m.Called()
	if sources := args.Get(0); sources != nil {
		if srcs, ok := sources.([]config.Source); ok {
			return srcs
		}
	}
	return nil
}

func (m *MockConfig) GetCommand() string {
	args := m.Called()
	if cmd := args.Get(0); cmd != nil {
		if str, ok := cmd.(string); ok {
			return str
		}
	}
	return "test"
}

// NewTestServerConfig creates a new server configuration for testing.
func NewTestServerConfig() *server.Config {
	return &server.Config{
		SecurityEnabled: false,
		APIKey:          "",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
}

// NewMockConfig creates a new mock configuration for testing.
func NewMockConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		Server: &server.Config{
			SecurityEnabled: false,
			APIKey:          "",
			Address:         ":8080",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
		},
		Crawler: &config.CrawlerConfig{
			MaxDepth:    3,
			RateLimit:   1 * time.Second,
			RandomDelay: 500 * time.Millisecond,
			Parallelism: 1,
			Sources:     []config.Source{},
		},
	}
}
