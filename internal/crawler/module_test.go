// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"testing"

	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
)

// mockSources implements sources.Interface for testing.
type mockSources struct {
	sources.Interface
}

// mockLogger implements logger.Interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...any) {}
func (m *mockLogger) Info(msg string, fields ...any)  {}
func (m *mockLogger) Warn(msg string, fields ...any)  {}
func (m *mockLogger) Error(msg string, fields ...any) {}
func (m *mockLogger) Fatal(msg string, fields ...any) {}
func (m *mockLogger) With(fields ...any) logger.Interface {
	return m
}

// TestCommonModule provides a test-specific common module that excludes the logger module.
var TestCommonModule = fx.Module("testCommon",
	// Suppress fx logging to reduce noise in the application logs.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),
	// Core modules used by most commands, excluding logger and sources.
	config.Module,
	logger.Module,
)

// TestConfigModule provides a test-specific config module that doesn't try to load files.
var TestConfigModule = fx.Module("testConfig",
	fx.Provide(
		fx.Annotate(
			func() config.Interface {
				mockCfg := &testutils.MockConfig{}
				mockCfg.On("GetAppConfig").Return(&config.AppConfig{
					Environment: "test",
					Name:        "gocrawl",
					Version:     "1.0.0",
					Debug:       true,
				})
				mockCfg.On("GetLogConfig").Return(&config.LogConfig{
					Level: "debug",
					Debug: true,
				})
				mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
					IndexName: "test-index",
				})
				mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				return mockCfg
			},
			fx.ResultTags(`name:"config"`),
		),
	),
)

// TestCrawlerModule provides a test version of the crawler module without common.Module
var TestCrawlerModule = fx.Module("crawler",
	article.Module,
	content.Module,
	fx.Provide(
		// Provide debugger
		func(logger common.Logger) debug.Debugger {
			return &debug.LogDebugger{
				Output: crawler.NewDebugLogger(logger),
			}
		},
		// Provide event bus
		events.NewBus,
		// Provide sources
		func() *sources.Sources {
			return &sources.Sources{}
		},
		// Article channel named instance
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, crawler.ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
		// Article index name
		fx.Annotate(
			func() string {
				return "articles"
			},
			fx.ResultTags(`name:"indexName"`),
		),
		// Content index name
		fx.Annotate(
			func() string {
				return "content"
			},
			fx.ResultTags(`name:"contentIndex"`),
		),
		// Article processor
		fx.Annotate(
			func(
				log common.Logger,
				articleService article.Interface,
				storage types.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) common.Processor {
				log.Debug("Providing article processor")
				return &article.ArticleProcessor{
					Logger:         log,
					ArticleService: articleService,
					Storage:        storage,
					IndexName:      params.IndexName,
					ArticleChan:    params.ArticleChan,
				}
			},
			fx.ResultTags(`name:"articleProcessor"`),
		),
		// Content processor
		fx.Annotate(
			func(
				log common.Logger,
				contentService content.Interface,
				storage types.Interface,
				params struct {
					fx.In
					IndexName string `name:"contentIndex"`
				},
			) common.Processor {
				log.Debug("Providing content processor")
				return content.NewContentProcessor(content.ProcessorParams{
					Logger:    log,
					Service:   contentService,
					Storage:   storage,
					IndexName: params.IndexName,
				})
			},
			fx.ResultTags(`name:"contentProcessor"`),
		),
		// Provide crawler
		crawler.ProvideCrawler,
	),
)

func setupTestApp() *fx.App {
	return fx.New(
		TestCommonModule,
		TestConfigModule,
		TestCrawlerModule,
		fx.Supply(mockSources{}),
		fx.NopLogger,
	)
}

// TestDependencyInjection verifies that all dependencies are properly injected into the Params struct.
func TestDependencyInjection(t *testing.T) {
	app := setupTestApp()
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify dependencies by invoking a function that checks them
	require.NoError(t, app.Err())
}

// TestModuleConstruction verifies that the crawler module can be constructed with all required dependencies.
func TestModuleConstruction(t *testing.T) {
	app := setupTestApp()
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestModuleLifecycle verifies that the crawler module can be started and stopped correctly.
func TestModuleLifecycle(t *testing.T) {
	app := setupTestApp()
	require.NoError(t, app.Start(t.Context()))
	require.NoError(t, app.Stop(t.Context()))
}

func TestModuleProvides(t *testing.T) {
	log := &mockLogger{}

	app := fxtest.New(t,
		fx.Supply(log),
		fx.Provide(
			fx.Annotate(func() logger.Interface { return log }, fx.As(new(logger.Interface))),
		),
		TestCrawlerModule,
	)

	app.RequireStart()
	app.RequireStop()
}
