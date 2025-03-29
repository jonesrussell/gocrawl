package crawl_test

import (
	"context"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// MockCrawler is a mock implementation of the crawler interface
type MockCrawler struct {
	mock.Mock
}

func (m *MockCrawler) Start(ctx context.Context, source string) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockCrawler) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCrawler) Wait() {
	m.Called()
}

func (m *MockCrawler) GetIndexManager() api.IndexManager {
	args := m.Called()
	if result := args.Get(0); result != nil {
		if im, ok := result.(api.IndexManager); ok {
			return im
		}
	}
	return nil
}

func (m *MockCrawler) Subscribe(handler events.Handler) {
	m.Called(handler)
}

func (m *MockCrawler) SetRateLimit(duration time.Duration) error {
	args := m.Called(duration)
	return args.Error(0)
}

func (m *MockCrawler) SetMaxDepth(depth int) {
	m.Called(depth)
}

func (m *MockCrawler) SetCollector(collector *colly.Collector) {
	m.Called(collector)
}

func (m *MockCrawler) GetMetrics() *collector.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	if metrics, ok := args.Get(0).(*collector.Metrics); ok {
		return metrics
	}
	return nil
}

// MockStorage is a mock implementation of the storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

func TestCommandExecution(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockCrawler := new(MockCrawler)
	mockStorage := new(MockStorage)
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)

	// Set up expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockCrawler.On("Start", mock.Anything, "test-source").Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()

	// Create test app
	app := fx.New(
		fx.Provide(
			func() crawler.Interface { return mockCrawler },
			func() types.Interface { return mockStorage },
			func() logger.Interface { return mockLogger },
			func() *signal.SignalHandler { return mockHandler },
			func() context.Context { return context.Background() },
			func() string { return "test-source" },
		),
		fx.Invoke(func(lc fx.Lifecycle, deps crawl.CommandDeps) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := deps.Storage.TestConnection(ctx); err != nil {
						return err
					}

					// Start crawler
					if err := deps.Crawler.Start(ctx, deps.SourceName); err != nil {
						return err
					}

					// Wait for crawler to complete
					go func() {
						deps.Crawler.Wait()
						deps.Handler.RequestShutdown()
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					return deps.Crawler.Stop(ctx)
				},
			})
		}),
	)

	// Start the app
	err := app.Start(context.Background())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(context.Background())
	require.NoError(t, err)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestCommandErrorHandling(t *testing.T) {
	t.Parallel()

	// Create root command with proper setup
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	cmd := crawl.Command()
	rootCmd.AddCommand(cmd)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "valid source name",
			args:    []string{"test-source"},
			wantErr: false,
		},
		{
			name:    "too many arguments",
			args:    []string{"test-source", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd.SetContext(t.Context())

			// Validate arguments against the crawl command
			err := cmd.ValidateArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandFlagHandling(t *testing.T) {
	t.Parallel()

	cmd := crawl.Command()
	ctx := context.Background()
	cmd.SetContext(ctx)

	// Test setting source flag
	err := cmd.Flags().Set("source", "test-source")
	require.NoError(t, err)

	// Verify flag value
	source, err := cmd.Flags().GetString("source")
	require.NoError(t, err)
	assert.Equal(t, "test-source", source)
}
