// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
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

func (m *MockConfig) GetAppConfig() *config.AppConfig {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		if appCfg, ok := cfg.(*config.AppConfig); ok {
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
		SourceFile:  "testdata/sources.yml",
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

// NewMockConfig creates a new mock configuration for testing.
func NewMockConfig() *config.Config {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "test",
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       true,
		},
		Log: config.LogConfig{
			Level: "debug",
			Debug: true,
		},
		Elasticsearch: config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test-index",
		},
		Server: config.ServerConfig{
			Address: ":8080",
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
				RateLimit: defaultRateLimit,
				CORS: struct {
					Enabled        bool     `yaml:"enabled"`
					AllowedOrigins []string `yaml:"allowed_origins"`
					AllowedMethods []string `yaml:"allowed_methods"`
					AllowedHeaders []string `yaml:"allowed_headers"`
					MaxAge         int      `yaml:"max_age"`
				}{
					Enabled:        true,
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders: []string{"*"},
					MaxAge:         defaultMaxAge,
				},
			},
		},
		Crawler: config.CrawlerConfig{
			MaxDepth:         defaultMaxDepth,
			RateLimit:        defaultRateLimit * time.Second,
			RandomDelay:      defaultRandomDelay,
			Parallelism:      defaultParallelism,
			IndexName:        "test-index",
			ContentIndexName: "test-content-index",
		},
	}
	return cfg
}
