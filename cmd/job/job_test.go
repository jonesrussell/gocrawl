package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/job"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storage "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/stretchr/testify/mock"
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
func (m *mockCrawler) GetMetrics() *collector.Metrics      { return nil }

// mockStorage implements storage.types.Interface for testing
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

// mockLogger implements logger.Interface for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

func (m *mockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

func (m *mockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure mockLogger implements logger.Interface
var _ logger.Interface = (*mockLogger)(nil)

func TestModuleProvides(t *testing.T) {
	mockLogger := &mockLogger{}
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLogger.On("Printf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Sync").Return(nil)

	mockCfg := &mockConfig{
		appConfig: &config.AppConfig{
			Environment: "test",
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       true,
		},
		crawlerConfig: &config.CrawlerConfig{
			MaxDepth:    2,
			RateLimit:   time.Second * 2,
			RandomDelay: time.Second,
			Parallelism: 2,
		},
		elasticsearchConfig: &config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test-index",
		},
		logConfig: &config.LogConfig{
			Level: "debug",
			Debug: true,
		},
		sources:      []config.Source{},
		serverConfig: &config.ServerConfig{Address: ":8080"},
		command:      "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() storage.Interface { return &mockStorage{} }, fx.As(new(storage.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}

func TestModuleLifecycle(t *testing.T) {
	mockLogger := &mockLogger{}
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLogger.On("Printf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Sync").Return(nil)

	mockCfg := &mockConfig{
		appConfig: &config.AppConfig{
			Environment: "test",
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       true,
		},
		crawlerConfig: &config.CrawlerConfig{
			MaxDepth:    2,
			RateLimit:   time.Second * 2,
			RandomDelay: time.Second,
			Parallelism: 2,
		},
		elasticsearchConfig: &config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test-index",
		},
		logConfig: &config.LogConfig{
			Level: "debug",
			Debug: true,
		},
		sources:      []config.Source{},
		serverConfig: &config.ServerConfig{Address: ":8080"},
		command:      "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() storage.Interface { return &mockStorage{} }, fx.As(new(storage.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}

func TestJobScheduling(t *testing.T) {
	mockLogger := &mockLogger{}
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLogger.On("Printf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Sync").Return(nil)

	mockCfg := &mockConfig{
		appConfig: &config.AppConfig{
			Environment: "test",
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       true,
		},
		crawlerConfig: &config.CrawlerConfig{
			MaxDepth:    2,
			RateLimit:   time.Second * 2,
			RandomDelay: time.Second,
			Parallelism: 2,
		},
		elasticsearchConfig: &config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test-index",
		},
		logConfig: &config.LogConfig{
			Level: "debug",
			Debug: true,
		},
		sources: []config.Source{
			{Name: "Test Source", URL: "https://test.com", RateLimit: time.Second, MaxDepth: 1},
		},
		serverConfig: &config.ServerConfig{Address: ":8080"},
		command:      "job",
	}

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(func() crawler.Interface { return &mockCrawler{} }, fx.As(new(crawler.Interface))),
			fx.Annotate(func() storage.Interface { return &mockStorage{} }, fx.As(new(storage.Interface))),
		),
		job.Module,
	)

	app.RequireStart()
	app.RequireStop()
}
