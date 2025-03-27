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

	// mockContentProcessor implements models.ContentProcessor for testing
	mockContentProcessor struct{}
)

// GetCrawlerConfig returns a mock crawler configuration with test values.
// This implementation provides basic crawler settings for testing purposes.
func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return &config.CrawlerConfig{
		MaxDepth:    3,
		Parallelism: 2,
		RateLimit:   time.Second,
	}
}

// GetElasticsearchConfig returns a mock Elasticsearch configuration with test values.
// This implementation provides a complete set of Elasticsearch settings including
// connection details, security settings, and retry policies.
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		APIKey:    "test-api-key",
		IndexName: "test-index",
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{
			ID:     "test-deployment",
			APIKey: "test-cloud-key",
		},
		TLS: struct {
			Enabled     bool   `yaml:"enabled"`
			SkipVerify  bool   `yaml:"skip_verify"`
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
			CA          string `yaml:"ca"`
		}{
			Enabled:    true,
			SkipVerify: true,
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     true,
			InitialWait: 1 * time.Second,
			MaxWait:     30 * time.Second,
			MaxRetries:  3,
		},
	}
}

// Search implements a mock search operation that returns an empty result set.
// This implementation is used for testing the search functionality without
// requiring a real Elasticsearch connection.
func (m *mockSearchManager) Search(_ context.Context, _ string, _ any) ([]any, error) {
	return []any{}, nil
}

// Count implements a mock count operation that returns zero.
// This implementation is used for testing count operations without
// requiring a real Elasticsearch connection.
func (m *mockSearchManager) Count(_ context.Context, _ string, _ any) (int64, error) {
	return 0, nil
}

// Aggregate implements a mock aggregate operation that returns an error.
// This implementation is used for testing error handling in aggregate operations.
func (m *mockSearchManager) Aggregate(_ context.Context, _ string, _ any) (any, error) {
	return nil, errors.New("aggregate not implemented in mock")
}

// Index implements a mock index operation that always succeeds.
// This implementation is used for testing index operations without
// requiring a real Elasticsearch connection.
func (m *mockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

// Close implements a mock close operation that always succeeds.
// This implementation is used for testing cleanup operations.
func (m *mockIndexManager) Close() error {
	return nil
}

// Store implements a mock store operation that always succeeds.
// This implementation is used for testing storage operations without
// requiring a real storage backend.
func (m *mockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

// Close implements a mock close operation that always succeeds.
// This implementation is used for testing cleanup operations.
func (m *mockStorage) Close() error {
	return nil
}

// GetSource returns a mock source configuration for testing.
// This implementation provides a complete set of source settings including
// URL, rate limiting, and content selectors.
func (m *mockSources) GetSource(_ string) (*sources.Config, error) {
	return &sources.Config{
		Name:      "test-source",
		URL:       "http://test.example.com",
		RateLimit: "1s",
		MaxDepth:  2,
		Selectors: sources.SelectorConfig{
			Title:       "h1",
			Description: "meta[name=description]",
			Content:     "article",
			Article: sources.ArticleSelectors{
				Container: "article",
				Title:     "h1",
				Body:      "article",
			},
		},
	}, nil
}

// ListSources returns a list of mock source configurations for testing.
// This implementation provides a single test source with complete configuration.
func (m *mockSources) ListSources() ([]*sources.Config, error) {
	return []*sources.Config{
		{
			Name:      "test-source",
			URL:       "http://test.example.com",
			RateLimit: "1s",
			MaxDepth:  2,
			Selectors: sources.SelectorConfig{
				Title:       "h1",
				Description: "meta[name=description]",
				Content:     "article",
				Article: sources.ArticleSelectors{
					Container: "article",
					Title:     "h1",
					Body:      "article",
				},
			},
		},
	}, nil
}

func (m *mockContentProcessor) Process(e *colly.HTMLElement) {
	// No-op implementation for testing
}
