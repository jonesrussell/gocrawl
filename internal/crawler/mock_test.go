// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"context"
	"errors"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Sentinel errors for testing
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoSources      = errors.New("no sources available")
	ErrNoResults      = errors.New("no results found")
	ErrNoAggregation  = errors.New("aggregation not implemented")
)

// Mock type declarations
type (
	// mockConfig implements config.Interface for testing.
	mockConfig struct {
		config.Interface
	}

	// mockSearchManager implements api.SearchManager for testing.
	mockSearchManager struct {
		api.SearchManager
	}

	// mockIndexManager implements api.IndexManager for testing.
	mockIndexManager struct {
		api.IndexManager
	}

	// mockStorage implements types.Interface for testing.
	mockStorage struct {
		types.Interface
	}

	// mockSources implements sources.Interface for testing.
	mockSources struct {
		sources.Interface
	}

	// mockContentProcessor implements collector.Processor for testing
	mockContentProcessor struct{}
)

// GetCrawlerConfig returns a mock crawler configuration with test values.
// This implementation provides basic crawler settings for testing purposes.
func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return &config.CrawlerConfig{
		BaseURL:          "http://test.com",
		MaxDepth:         2,
		RateLimit:        time.Second,
		RandomDelay:      time.Second,
		IndexName:        "test-index",
		ContentIndexName: "test-content-index",
		SourceFile:       "sources.yml",
		Parallelism:      1,
	}
}

// GetElasticsearchConfig returns a mock Elasticsearch configuration.
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "test",
		Password:  "test",
	}
}

// GetLogConfig returns a mock logging configuration.
func (m *mockConfig) GetLogConfig() *config.LogConfig {
	return &config.LogConfig{
		Level: "debug",
	}
}

// GetAppConfig returns a mock application configuration.
func (m *mockConfig) GetAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Environment: "test",
		Name:        "test-app",
		Version:     "1.0.0",
		Debug:       true,
	}
}

// GetSources returns a mock list of sources.
func (m *mockConfig) GetSources() []config.Source {
	return []config.Source{
		{
			Name:     "test-source",
			URL:      "http://test.com",
			MaxDepth: 2,
		},
	}
}

// GetServerConfig returns a mock server configuration.
func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Address:      ":8080",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}
}

// GetCommand returns a mock command.
func (m *mockConfig) GetCommand() string {
	return "test"
}

// Mock implementations for other interfaces
func (m *mockSearchManager) Search(_ context.Context, _ string, _ any) ([]any, error) {
	return nil, ErrNoResults
}

func (m *mockSearchManager) Count(_ context.Context, _ string, _ any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Aggregate(_ context.Context, _ string, _ any) (any, error) {
	return nil, ErrNoAggregation
}

func (m *mockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockIndexManager) Close() error {
	return nil
}

func (m *mockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockSources) GetSource(_ string) (*sources.Config, error) {
	return nil, ErrNotImplemented
}

func (m *mockSources) ListSources() ([]*sources.Config, error) {
	return nil, ErrNoSources
}

func (m *mockContentProcessor) Process(e *colly.HTMLElement) error {
	// No-op for testing
	return nil
}
