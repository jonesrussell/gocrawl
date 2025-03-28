// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
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
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the crawl module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

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
		// Core dependencies
		collector.Module(),

		// Provide all required dependencies
		fx.Provide(
			// Command-specific dependencies
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return "Test Source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() chan struct{} { return make(chan struct{}) },
				fx.ResultTags(`name:"commandDone"`),
			),
			fx.Annotate(
				func() chan *models.Article {
					return make(chan *models.Article, crawler.ArticleChannelBufferSize)
				},
				fx.ResultTags(`name:"articleChannel"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return signal.NewSignalHandler(mockLogger) },
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Test dependencies
			fx.Annotate(
				func() struct {
					fx.Out
					Logger     logger.Interface  `name:"logger"`
					Config     config.Interface  `name:"testConfig"`
					Sources    sources.Interface `name:"testSourceManager"`
					ArticleSvc article.Interface `name:"testArticleService"`
					IndexMgr   api.IndexManager  `name:"indexManager"`
					Content    content.Interface
					Crawler    crawler.Interface
				} {
					return struct {
						fx.Out
						Logger     logger.Interface  `name:"logger"`
						Config     config.Interface  `name:"testConfig"`
						Sources    sources.Interface `name:"testSourceManager"`
						ArticleSvc article.Interface `name:"testArticleService"`
						IndexMgr   api.IndexManager  `name:"indexManager"`
						Content    content.Interface
						Crawler    crawler.Interface
					}{
						Logger:     mockLogger,
						Config:     mockCfg,
						Sources:    testSources,
						ArticleSvc: article.NewService(mockLogger, config.DefaultArticleSelectors()),
						IndexMgr:   testutils.NewMockIndexManager(),
						Content:    content.NewService(mockLogger),
						Crawler:    testutils.NewMockCrawler(),
					}
				},
			),
		),

		// Decorate storage-related dependencies with mocks
		fx.Decorate(
			func() types.Interface { return testutils.NewMockStorage() },
			func() api.SearchManager { return testutils.NewMockSearchManager() },
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
	mockLogger := &testutils.MockLogger{}
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

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
		// Core dependencies
		collector.Module(),

		// Provide all required dependencies
		fx.Provide(
			// Command-specific dependencies
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return "Test Source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() chan struct{} { return make(chan struct{}) },
				fx.ResultTags(`name:"commandDone"`),
			),
			fx.Annotate(
				func() chan *models.Article {
					return make(chan *models.Article, crawler.ArticleChannelBufferSize)
				},
				fx.ResultTags(`name:"articleChannel"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return signal.NewSignalHandler(mockLogger) },
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Test dependencies
			fx.Annotate(
				func() struct {
					fx.Out
					Logger     logger.Interface  `name:"logger"`
					Config     config.Interface  `name:"testConfig"`
					Sources    sources.Interface `name:"testSourceManager"`
					ArticleSvc article.Interface `name:"testArticleService"`
					IndexMgr   api.IndexManager  `name:"indexManager"`
					Content    content.Interface
					Crawler    crawler.Interface
				} {
					return struct {
						fx.Out
						Logger     logger.Interface  `name:"logger"`
						Config     config.Interface  `name:"testConfig"`
						Sources    sources.Interface `name:"testSourceManager"`
						ArticleSvc article.Interface `name:"testArticleService"`
						IndexMgr   api.IndexManager  `name:"indexManager"`
						Content    content.Interface
						Crawler    crawler.Interface
					}{
						Logger:     mockLogger,
						Config:     mockCfg,
						Sources:    testSources,
						ArticleSvc: article.NewService(mockLogger, config.DefaultArticleSelectors()),
						IndexMgr:   testutils.NewMockIndexManager(),
						Content:    content.NewService(mockLogger),
						Crawler:    testutils.NewMockCrawler(),
					}
				},
			),
		),

		// Decorate storage-related dependencies with mocks
		fx.Decorate(
			func() types.Interface { return testutils.NewMockStorage() },
			func() api.SearchManager { return testutils.NewMockSearchManager() },
		),
	)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule,
	)

	require.NoError(t, app.Err())
}
