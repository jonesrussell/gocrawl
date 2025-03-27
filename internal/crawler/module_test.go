// Package crawler_test provides tests for the crawler package.
// It verifies the dependency injection setup and ensures all required components
// are properly wired together using the fx framework.
package crawler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig implements config.Interface for testing.
// It provides mock implementations of configuration methods with predefined test values.
type mockConfig struct {
	config.Interface
}

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
		APIKey:    "test-api-key", // Required API key
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

// mockSearchManager implements api.SearchManager for testing.
// It provides mock implementations of search operations that return empty results.
type mockSearchManager struct {
	api.SearchManager
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

// mockIndexManager implements api.IndexManager for testing.
// It provides mock implementations of index operations that always succeed.
type mockIndexManager struct {
	api.IndexManager
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

// mockStorage implements types.Interface for testing.
// It provides mock implementations of storage operations that always succeed.
type mockStorage struct {
	types.Interface
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

// mockSources implements sources.Interface for testing.
// It provides mock implementations of source management operations with predefined test data.
type mockSources struct {
	sources.Interface
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

// TestModule tests that the crawler module provides all necessary dependencies.
// It verifies that the module can be constructed without errors and that all
// required dependencies are available.
func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock instances for all required dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockIndex := &mockIndexManager{}
	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockStorage := &mockStorage{}

	// Set up debug logging expectations for the mock logger
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Verify that mockLogger implements the required interface
	var _ logger.Interface = mockLogger

	// Create a new fx application with the test dependencies
	app := fxtest.New(t,
		fx.Provide(
			// Provide mock logger with the correct name
			fx.Annotate(
				func() logger.Interface {
					mockLogger.Debug("Providing test logger")
					return mockLogger
				},
				fx.ResultTags(`name:"testLogger"`),
			),
			// Provide mock config with the correct name
			fx.Annotate(
				func() config.Interface {
					mockLogger.Debug("Providing test config")
					return mockCfg
				},
				fx.ResultTags(`name:"testConfig"`),
			),
			// Provide mock index manager with the correct name
			fx.Annotate(
				func() api.IndexManager {
					mockLogger.Debug("Providing test index manager")
					return mockIndex
				},
				fx.ResultTags(`name:"testIndexManager"`),
			),
			// Provide mock search manager with the correct name
			fx.Annotate(
				func() api.SearchManager {
					mockLogger.Debug("Providing test search manager")
					return mockSearchManager
				},
				fx.ResultTags(`name:"testSearchManager"`),
			),
			// Provide mock storage with the correct name
			fx.Annotate(
				func() types.Interface {
					mockLogger.Debug("Providing test storage")
					return mockStorage
				},
				fx.ResultTags(`name:"testStorage"`),
			),
		),
		crawler.Module,
	)
	require.NoError(t, app.Err())
}

// TestModuleProvides tests that the crawler module provides all necessary dependencies.
// It verifies that the module can be started and stopped without errors, ensuring
// that all lifecycle hooks work correctly.
func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock instances for all required dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockIndex := &mockIndexManager{}

	// Set up debug logging expectations for the mock logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create a new fx application with the test dependencies
	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			// Provide mock logger with the correct name
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.ResultTags(`name:"testLogger"`),
			),
			// Provide mock config with the correct name
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.ResultTags(`name:"testConfig"`),
			),
			// Provide mock storage with the correct name
			fx.Annotate(
				func() types.Interface { return mockStore },
				fx.ResultTags(`name:"testStorage"`),
			),
			// Provide mock index manager with the correct name
			fx.Annotate(
				func() api.IndexManager { return mockIndex },
				fx.ResultTags(`name:"testIndexManager"`),
			),
		),
		crawler.Module,
	)

	// Verify that the application can be started and stopped
	app.RequireStart()
	app.RequireStop()
}

// TestAppDependencies tests that all dependencies are properly wired up in the crawler module.
// It verifies that all required dependencies are available and properly initialized when
// constructing the CrawlDeps struct.
func TestAppDependencies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock instances for all required dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockIndex := &mockIndexManager{}
	mockSearchManager := &mockSearchManager{}
	mockSources := &mockSources{}
	mockSignalHandler := signal.NewSignalHandler(mockLogger)

	// Set up debug logging expectations for the mock logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create a new fx application with the test dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		crawler.Module,
		fx.Provide(
			// Provide mock logger with the correct name
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.ResultTags(`name:"testLogger"`),
			),
			// Provide mock config with the correct name
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.ResultTags(`name:"testConfig"`),
			),
			// Provide mock storage with the correct name
			fx.Annotate(
				func() types.Interface { return mockStore },
				fx.ResultTags(`name:"testStorage"`),
			),
			// Provide mock index manager with the correct name
			fx.Annotate(
				func() api.IndexManager { return mockIndex },
				fx.ResultTags(`name:"testIndexManager"`),
			),
			// Provide mock search manager with the correct name
			fx.Annotate(
				func() api.SearchManager { return mockSearchManager },
				fx.ResultTags(`name:"testSearchManager"`),
			),
			// Provide mock sources with the correct name
			fx.Annotate(
				func() sources.Interface { return mockSources },
				fx.ResultTags(`name:"sourceManager"`),
			),
			// Provide source name with the correct name
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			// Provide mock signal handler with the correct name
			fx.Annotate(
				func() *signal.SignalHandler { return mockSignalHandler },
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Provide context with the correct name
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			// Provide article channel with the correct name
			fx.Annotate(
				func() chan *models.Article { return make(chan *models.Article) },
				fx.ResultTags(`name:"articleChannel"`),
			),
		),
		// Verify that all dependencies are properly injected
		fx.Invoke(func(deps crawler.CrawlDeps) {
			assert.NotNil(t, deps.Logger)
			assert.NotNil(t, deps.Config)
			assert.NotNil(t, deps.Storage)
			assert.NotNil(t, deps.Crawler)
			assert.NotNil(t, deps.Processors)
			assert.NotNil(t, deps.Done)
			assert.NotNil(t, deps.Context)
			assert.NotNil(t, deps.Sources)
			assert.NotNil(t, deps.SourceName)
			assert.NotNil(t, deps.Handler)
		}),
	)

	// Verify that the application can be started and stopped
	app.RequireStart()
	app.RequireStop()
}
