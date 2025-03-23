// Package job_test implements tests for the job scheduler command.
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
	"github.com/jonesrussell/gocrawl/internal/storage"
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

func (m *mockConfig) GetAppConfig() *config.AppConfig {
	return m.appConfig
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return m.crawlerConfig
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return m.elasticsearchConfig
}

func (m *mockConfig) GetLogConfig() *config.LogConfig {
	return m.logConfig
}

func (m *mockConfig) GetSources() []config.Source {
	return m.sources
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	return m.serverConfig
}

func (m *mockConfig) GetCommand() string {
	return m.command
}

// mockCrawler implements crawler.Interface for testing
type mockCrawler struct{}

func (m *mockCrawler) Start(ctx context.Context, baseURL string) error {
	return nil
}

func (m *mockCrawler) Stop(ctx context.Context) error {
	return nil
}

func (m *mockCrawler) Subscribe(handler events.Handler) {}

func (m *mockCrawler) SetRateLimit(duration string) error {
	return nil
}

func (m *mockCrawler) SetMaxDepth(depth int) {}

func (m *mockCrawler) SetCollector(collector *colly.Collector) {}

func (m *mockCrawler) GetIndexManager() api.IndexManager {
	return nil
}

func (m *mockCrawler) Wait() {}

// mockStorage implements storage.Interface for testing
type mockStorage struct{}

func (m *mockStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	return nil
}

func (m *mockStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	return nil
}

func (m *mockStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	return nil
}

func (m *mockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	return nil
}

func (m *mockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	return nil
}

func (m *mockStorage) DeleteIndex(ctx context.Context, index string) error {
	return nil
}

func (m *mockStorage) ListIndices(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	return nil, errors.New("mapping not found")
}

func (m *mockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	return nil
}

func (m *mockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	return false, nil
}

func (m *mockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	return nil, nil
}

func (m *mockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	return "", errors.New("index health not available")
}

func (m *mockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	return 0, nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *mockStorage) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	return nil, errors.New("aggregation not supported in mock")
}

func (m *mockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	return 0, nil
}

// TestModuleProvides tests that the job module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	// Create a mock logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)

	// Create a mock config with required fields
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources:             []config.Source{},
		serverConfig:        &config.ServerConfig{},
		command:             "job",
	}

	// Create mock implementations
	mockCrawlerInstance := &mockCrawler{}
	mockStorageInstance := &mockStorage{}

	// Create a test app with the mock dependencies
	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(
				func() crawler.Interface { return mockCrawlerInstance },
				fx.As(new(crawler.Interface)),
			),
			fx.Annotate(
				func() storage.Interface { return mockStorageInstance },
				fx.As(new(storage.Interface)),
			),
		),
		job.Module,
	)

	// Test lifecycle
	app.RequireStart()
	app.RequireStop()
}

// TestModuleLifecycle tests the lifecycle hooks of the job module
func TestModuleLifecycle(t *testing.T) {
	// Create a mock logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)

	// Create a mock config with required fields
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources:             []config.Source{},
		serverConfig:        &config.ServerConfig{},
		command:             "job",
	}

	// Create mock implementations
	mockCrawlerInstance := &mockCrawler{}
	mockStorageInstance := &mockStorage{}

	// Create a test app with the mock dependencies
	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(
				func() crawler.Interface { return mockCrawlerInstance },
				fx.As(new(crawler.Interface)),
			),
			fx.Annotate(
				func() storage.Interface { return mockStorageInstance },
				fx.As(new(storage.Interface)),
			),
		),
		job.Module,
	)

	// Test lifecycle
	app.RequireStart()
	app.RequireStop()
}

// TestJobScheduling tests the job scheduling functionality
func TestJobScheduling(t *testing.T) {
	// Create a mock logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockInterface(ctrl)

	// Create a mock config with required fields
	mockCfg := &mockConfig{
		appConfig:           &config.AppConfig{},
		crawlerConfig:       &config.CrawlerConfig{},
		elasticsearchConfig: &config.ElasticsearchConfig{},
		logConfig:           &config.LogConfig{},
		sources: []config.Source{
			{
				Name:      "Test Source",
				URL:       "https://test.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
		},
		serverConfig: &config.ServerConfig{},
		command:      "job",
	}

	// Create mock implementations
	mockCrawlerInstance := &mockCrawler{}
	mockStorageInstance := &mockStorage{}

	// Create a test app with the mock dependencies
	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(
				func() crawler.Interface { return mockCrawlerInstance },
				fx.As(new(crawler.Interface)),
			),
			fx.Annotate(
				func() storage.Interface { return mockStorageInstance },
				fx.As(new(storage.Interface)),
			),
		),
		job.Module,
	)

	// Test job scheduling
	app.RequireStart()
	app.RequireStop()
}
