// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	sourcestest "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}

func (m *mockSearchManager) Search(_ context.Context, _ string, _ any) ([]any, error) {
	return []any{}, nil
}

func (m *mockSearchManager) Count(_ context.Context, _ string, _ any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Aggregate(_ context.Context, _ string, _ any) (any, error) {
	return nil, errors.New("aggregate not implemented in mock")
}

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	api.IndexManager
}

func (m *mockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockIndexManager) Close() error {
	return nil
}

// mockStorage implements types.Interface for testing
type mockStorage struct {
	types.Interface
}

func (m *mockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

// TestModuleProvides tests that the crawl module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockCfg := configtest.NewMockConfig().WithSources([]config.Source{
		{
			Name: "Test Source",
			URL:  "http://test.example.com",
		},
	})

	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "http://test.example.com",
			RateLimit: time.Second,
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	// Create a test-specific module that excludes config.Module and sources.Module
	testModule := fx.Module("test",
		// Core dependencies (excluding config, logger, and storage modules)
		content.Module,
		collector.Module(),
		crawler.Module,

		// Provide all required dependencies
		fx.Provide(
			func() context.Context { return t.Context() },
			// Test dependencies
			fx.Annotate(
				func() struct {
					fx.Out
					Logger     logger.Interface  `name:"testLogger"`
					Config     config.Interface  `name:"testConfig"`
					Sources    sources.Interface `name:"sourceManager"`
					SourceName string            `name:"sourceName"`
					ArticleSvc article.Interface `name:"testArticleService"`
				} {
					return struct {
						fx.Out
						Logger     logger.Interface  `name:"testLogger"`
						Config     config.Interface  `name:"testConfig"`
						Sources    sources.Interface `name:"sourceManager"`
						SourceName string            `name:"sourceName"`
						ArticleSvc article.Interface `name:"testArticleService"`
					}{
						Logger:     mockLogger,
						Config:     mockCfg,
						Sources:    testSources,
						SourceName: "Test Source",
						ArticleSvc: article.NewService(mockLogger, config.DefaultArticleSelectors()),
					}
				},
			),
		),

		// Decorate storage-related dependencies with mocks
		fx.Decorate(
			func() types.Interface { return &mockStorage{} },
			func() api.IndexManager { return &mockIndexManager{} },
			func() api.SearchManager { return &mockSearchManager{} },
		),
	)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule,
	)

	require.NoError(t, app.Err())
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create test dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockCfg := configtest.NewMockConfig().
		WithSources([]config.Source{
			{
				Name: "Test Source",
				URL:  "https://test.com",
			},
		}).
		WithCrawlerConfig(&config.CrawlerConfig{
			MaxDepth:    3,
			Parallelism: 2,
		}).
		WithElasticsearchConfig(&config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			APIKey:    "test-api-key",
			IndexName: "test-index",
		})

	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "https://test.com",
			RateLimit: time.Second,
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	// Create a test-specific module that excludes config.Module and sources.Module
	testModule := fx.Module("test",
		// Core dependencies (excluding config, logger, and storage modules)
		content.Module,
		collector.Module(),
		crawler.Module,

		// Provide all required dependencies
		fx.Provide(
			func() context.Context { return t.Context() },
			// Test dependencies
			fx.Annotate(
				func() struct {
					fx.Out
					Logger     logger.Interface  `name:"testLogger"`
					Config     config.Interface  `name:"testConfig"`
					Sources    sources.Interface `name:"sourceManager"`
					SourceName string            `name:"sourceName"`
					ArticleSvc article.Interface `name:"testArticleService"`
					IndexMgr   api.IndexManager  `name:"testIndexManager"`
				} {
					return struct {
						fx.Out
						Logger     logger.Interface  `name:"testLogger"`
						Config     config.Interface  `name:"testConfig"`
						Sources    sources.Interface `name:"sourceManager"`
						SourceName string            `name:"sourceName"`
						ArticleSvc article.Interface `name:"testArticleService"`
						IndexMgr   api.IndexManager  `name:"testIndexManager"`
					}{
						Logger:     mockLogger,
						Config:     mockCfg,
						Sources:    testSources,
						SourceName: "Test Source",
						ArticleSvc: article.NewService(mockLogger, config.DefaultArticleSelectors()),
						IndexMgr:   &mockIndexManager{},
					}
				},
			),
		),

		// Decorate storage-related dependencies with mocks
		fx.Decorate(
			func() types.Interface { return &mockStorage{} },
			func() api.SearchManager { return &mockSearchManager{} },
		),
	)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule,
	)

	require.NoError(t, app.Err())
}
