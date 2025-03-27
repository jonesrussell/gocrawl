// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements logger.Interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	currentTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *MockTimeProvider) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

// MockSecurityMiddleware implements SecurityMiddlewareInterface for testing
type MockSecurityMiddleware struct {
	mock.Mock
}

func (m *MockSecurityMiddleware) Cleanup(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockSecurityMiddleware) WaitCleanup() {
	m.Called()
}

func (m *MockSecurityMiddleware) Middleware() gin.HandlerFunc {
	args := m.Called()
	if fn := args.Get(0); fn != nil {
		return fn.(gin.HandlerFunc)
	}
	return func(c *gin.Context) { c.Next() }
}

// MockSearchManager implements SearchManager interface for testing
type MockSearchManager struct {
	mock.Mock
}

func (m *MockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

func (m *MockSearchManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

func (m *MockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockConfig implements config.Interface for testing
type MockConfig struct {
	mock.Mock
	configFile string
}

func (m *MockConfig) GetAppConfig() *config.AppConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*config.AppConfig)
	}
	return nil
}

func (m *MockConfig) GetLogConfig() *config.LogConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*config.LogConfig)
	}
	return nil
}

func (m *MockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return &config.CrawlerConfig{
		SourceFile:  "testdata/sources.yml",
		MaxDepth:    2,
		RateLimit:   1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
		Parallelism: 2,
	}
}

func (m *MockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*config.ElasticsearchConfig)
	}
	return nil
}

func (m *MockConfig) GetServerConfig() *config.ServerConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*config.ServerConfig)
	}
	return nil
}

// NewTestServerConfig creates a new ServerConfig for testing
func NewTestServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Security: struct {
			Enabled   bool   `yaml:"enabled"`
			APIKey    string `yaml:"api_key"`
			RateLimit int    `yaml:"rate_limit"`
			CORS      struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			} `yaml:"cors"`
			TLS struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			} `yaml:"tls"`
		}{
			Enabled:   true,
			APIKey:    "test-key",
			RateLimit: 100,
			CORS: struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			}{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Authorization", "X-API-Key"},
				MaxAge:         86400,
			},
		},
	}
}

// NewMockConfig creates a new mock config that loads from the test config file
func NewMockConfig() *MockConfig {
	return &MockConfig{
		configFile: "testdata/config.yml",
	}
}
