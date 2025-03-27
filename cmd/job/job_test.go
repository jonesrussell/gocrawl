package job_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig implements config.Interface for testing
type mockConfig struct {
	appConfig           *config.AppConfig
	crawlerConfig       *config.CrawlerConfig
	elasticsearchConfig *config.ElasticsearchConfig
	logConfig           *config.LogConfig
	sources             []config.Source
	serverConfig        *config.ServerConfig
	command             string
}

func (m *mockConfig) GetAppConfig() *config.AppConfig         { return m.appConfig }
func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig { return m.crawlerConfig }
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return m.elasticsearchConfig
}
func (m *mockConfig) GetLogConfig() *config.LogConfig       { return m.logConfig }
func (m *mockConfig) GetSources() []config.Source           { return m.sources }
func (m *mockConfig) GetServerConfig() *config.ServerConfig { return m.serverConfig }
func (m *mockConfig) GetCommand() string                    { return m.command }

// mockCrawler implements crawler.Interface for testing
type mockCrawler struct{}

func (m *mockCrawler) Start(context.Context, string) error { return nil }
func (m *mockCrawler) Stop(context.Context) error          { return nil }
func (m *mockCrawler) Subscribe(events.Handler)            {}
func (m *mockCrawler) SetRateLimit(time.Duration) error    { return nil }
func (m *mockCrawler) SetMaxDepth(_ int)                   {}
func (m *mockCrawler) SetCollector(_ *colly.Collector)     {}
func (m *mockCrawler) GetIndexManager() api.IndexManager   { return nil }
func (m *mockCrawler) Wait()                               {}
func (m *mockCrawler) GetMetrics() *crawler.Metrics        { return nil }

// mockStorage implements types.Interface for testing
type mockStorage struct{}

func (m *mockStorage) IndexDocument(context.Context, string, string, any) error  { return nil }
func (m *mockStorage) GetDocument(context.Context, string, string, any) error    { return nil }
func (m *mockStorage) DeleteDocument(context.Context, string, string) error      { return nil }
func (m *mockStorage) BulkIndex(context.Context, string, []any) error            { return nil }
func (m *mockStorage) CreateIndex(context.Context, string, map[string]any) error { return nil }
func (m *mockStorage) DeleteIndex(context.Context, string) error                 { return nil }
func (m *mockStorage) ListIndices(context.Context) ([]string, error)             { return nil, nil }
func (m *mockStorage) GetMapping(context.Context, string) (map[string]any, error) {
	return nil, errors.New("mapping not found")
}
func (m *mockStorage) UpdateMapping(context.Context, string, map[string]any) error { return nil }
func (m *mockStorage) IndexExists(context.Context, string) (bool, error)           { return false, nil }
func (m *mockStorage) Search(context.Context, string, any) ([]any, error)          { return nil, nil }
func (m *mockStorage) GetIndexHealth(context.Context, string) (string, error) {
	return "", errors.New("index health not available")
}
func (m *mockStorage) GetIndexDocCount(context.Context, string) (int64, error) { return 0, nil }
func (m *mockStorage) Close() error                                            { return nil }
func (m *mockStorage) Ping(context.Context) error                              { return nil }
func (m *mockStorage) TestConnection(context.Context) error                    { return nil }
func (m *mockStorage) Aggregate(context.Context, string, any) (any, error) {
	return nil, errors.New("aggregation not supported in mock")
}
func (m *mockStorage) Count(context.Context, string, any) (int64, error) { return 0, nil }

func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources:             []config.Source{},
		serverConfig:        &config.ServerConfig{},
		command:             "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() types.Interface { return &mockStorage{} }, fx.As(new(types.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}

func TestModuleLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources:             []config.Source{},
		serverConfig:        &config.ServerConfig{},
		command:             "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() types.Interface { return &mockStorage{} }, fx.As(new(types.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}

func TestJobScheduling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources: []config.Source{
			{Name: "Test Source", URL: "https://test.com", RateLimit: time.Second, MaxDepth: 1},
		},
		serverConfig: &config.ServerConfig{},
		command:      "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() types.Interface { return &mockStorage{} }, fx.As(new(types.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}
