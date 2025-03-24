// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	sourcestest "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}

func (m *mockSearchManager) Search(context.Context, string, any) ([]any, error) {
	return []any{}, nil
}

func (m *mockSearchManager) Count(context.Context, string, any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Aggregate(context.Context, string, any) (any, error) {
	return nil, errors.New("aggregate not implemented in mock")
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
			RateLimit: "1s",
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	// Create a test-specific module that excludes config.Module
	testModule := fx.Module("test",
		// Core dependencies (excluding config and logger modules)
		sources.Module,
		api.Module,

		// Feature modules
		article.Module,
		content.Module,
		collector.Module(),
		crawler.Module,

		// Provide all required dependencies
		fx.Provide(
			// Logger
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
			// Config
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.As(new(config.Interface)),
			),
			// Index Manager
			fx.Annotate(
				func() api.IndexManager { return nil },
				fx.As(new(api.IndexManager)),
			),
			// Search Manager
			fx.Annotate(
				func() api.SearchManager { return &mockSearchManager{} },
				fx.As(new(api.SearchManager)),
			),
			// Sources
			func() *sources.Sources { return testSources },
			// Named dependencies
			fx.Annotate(
				func() string { return "Test Source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func(sources sources.Interface) (string, string) {
					return "test_content", "test_articles"
				},
				fx.ParamTags(`name:"sourceManager"`),
				fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
			),
			func() chan *models.Article {
				return make(chan *models.Article, 100)
			},
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
		})

	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "https://test.com",
			RateLimit: "1s",
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	var crawlerInstance crawler.Interface

	// Create a test-specific module that excludes config.Module
	testModule := fx.Module("test",
		// Core dependencies (excluding config and logger modules)
		sources.Module,
		api.Module,

		// Feature modules
		article.Module,
		content.Module,
		collector.Module(),
		crawler.Module,

		// Provide all required dependencies
		fx.Provide(
			// Logger
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
			// Config
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.As(new(config.Interface)),
			),
			// Index Manager
			fx.Annotate(
				func() api.IndexManager { return api.NewMockIndexManager() },
				fx.As(new(api.IndexManager)),
			),
			// Search Manager
			fx.Annotate(
				func() api.SearchManager { return &mockSearchManager{} },
				fx.As(new(api.SearchManager)),
			),
			// Sources
			func() *sources.Sources { return testSources },
			// Named dependencies
			fx.Annotate(
				func() string { return "Test Source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func(sources sources.Interface) (string, string) {
					return "test_content", "test_articles"
				},
				fx.ParamTags(`name:"sourceManager"`),
				fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
			),
			func() chan *models.Article {
				return make(chan *models.Article, 100)
			},
		),
	)

	// Create test app with test-specific module
	app := fxtest.New(t,
		fx.NopLogger,
		testModule,
		fx.Populate(&crawlerInstance),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer func(app *fxtest.App, ctx context.Context) {
		err := app.Stop(ctx)
		if err != nil {
			t.Errorf("Error stopping app: %v", err)
		}
	}(app, t.Context())

	// Verify crawler configuration
	assert.NotNil(t, crawlerInstance, "Crawler should be provided")
	// Note: Add more specific crawler configuration checks here once crawler exposes them
}
