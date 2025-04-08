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

func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if serverCfg, ok := cfg.(*config.ServerConfig); ok {
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

// NewTestServerConfig creates a new ServerConfig for testing
func NewTestServerConfig() *server.Config {
	return &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:test-key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
}

// NewMockConfig creates a new mock configuration for testing.
func NewMockConfig() *config.Config {
	appCfg := app.New(
		app.WithEnvironment("test"),
		app.WithName("gocrawl"),
		app.WithVersion("1.0.0"),
		app.WithDebug(true),
	)

	logCfg := &config.LogConfig{
		Level: "debug",
	}

	elasticCfg := &config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
	}

	serverCfg := &server.Config{
		SecurityEnabled: false,
		APIKey:          "",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}

	crawlerCfg := &config.CrawlerConfig{
		MaxDepth:    2,
		RateLimit:   time.Second * 5,
		Parallelism: 4,
	}

	return &config.Config{
		App:           appCfg,
		Log:           logCfg,
		Elasticsearch: elasticCfg,
		Server:        serverCfg,
		Crawler:       crawlerCfg,
	}
}
