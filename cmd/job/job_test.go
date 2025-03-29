// Package job_test provides tests for the job command package.
package job_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	err                 error
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

func TestJobCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(*testing.T) (*mockLogger, *mockConfig)
		expectedError  string
		expectedLogs   []string
		timeout        time.Duration
		shouldComplete bool
	}{
		{
			name: "successful_initialization",
			setup: func(t *testing.T) (*mockLogger, *mockConfig) {
				mockLogger := &mockLogger{}
				mockConfig := &mockConfig{}
				return mockLogger, mockConfig
			},
			expectedError: "",
			timeout:       5 * time.Second,
		},
		{
			name: "config_error_handling",
			setup: func(t *testing.T) (*mockLogger, *mockConfig) {
				mockLogger := &mockLogger{}
				mockConfig := &mockConfig{
					err: errors.New("config error"),
				}
				return mockLogger, mockConfig
			},
			expectedError: "failed to create config: config error",
			timeout:       5 * time.Second,
		},
		{
			name: "graceful_shutdown",
			setup: func(t *testing.T) (*mockLogger, *mockConfig) {
				mockLogger := &mockLogger{}
				mockConfig := &mockConfig{}
				mockLogger.On("Info", "Context cancelled, initiating shutdown").Return()
				mockLogger.On("Info", "Job completed").Return()
				return mockLogger, mockConfig
			},
			expectedError: "",
			timeout:       5 * time.Second,
		},
		{
			name: "shutdown_timeout",
			setup: func(t *testing.T) (*mockLogger, *mockConfig) {
				mockLogger := &mockLogger{}
				mockConfig := &mockConfig{}
				mockLogger.On("Info", "Context cancelled, initiating shutdown").Return()
				mockLogger.On("Info", "Job completed").Return()
				return mockLogger, mockConfig
			},
			expectedError: "",
			timeout:       5 * time.Second,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Set up mocks
			mockLogger, mockConfig := tt.setup(t)

			// Create command with test dependencies
			cmd := job.NewJobCommand(job.JobCommandDeps{
				Logger: mockLogger,
				Config: mockConfig,
			})

			// Set up test command context
			cmd.SetContext(ctx)

			// Execute command in a goroutine
			errChan := make(chan error, 1)
			go func() {
				errChan <- cmd.Execute()
			}()

			// Wait for either error or timeout
			select {
			case err := <-errChan:
				if tt.expectedError != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.expectedError)
				} else {
					require.NoError(t, err)
				}
			case <-ctx.Done():
				t.Fatal("test timed out")
			}

			// Verify mock expectations
			mockLogger.AssertExpectations(t)
		})
	}
}
