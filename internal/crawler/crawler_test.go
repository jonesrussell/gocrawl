package crawler_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	mock.Mock
}

func (m *mockIndexManager) CreateIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockIndexManager) EnsureIndex(ctx context.Context, name string, settings any) error {
	args := m.Called(ctx, name, settings)
	return args.Error(0)
}

func (m *mockIndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// MockProcessor implements common.Processor for testing
type MockProcessor struct {
	mock.Mock
	ProcessCalls int
}

// ProcessJob implements common.Processor
func (m *MockProcessor) ProcessJob(ctx context.Context, job *common.Job) {
	m.Called(ctx, job)
}

// ProcessHTML implements common.Processor
func (m *MockProcessor) ProcessHTML(e *colly.HTMLElement) error {
	m.ProcessCalls++
	args := m.Called(e)
	return args.Error(0)
}

// GetMetrics implements common.Processor
func (m *MockProcessor) GetMetrics() *common.Metrics {
	args := m.Called()
	return args.Get(0).(*common.Metrics)
}

// Ensure MockProcessor implements common.Processor
var _ common.Processor = (*MockProcessor)(nil)

func TestCrawlerStartup(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create mock sources
	mockSources := &sources.Sources{}
	mockSources.SetSources([]sources.Config{
		{
			Name:      "test",
			URL:       "http://test.com",
			RateLimit: time.Hour,
			MaxDepth:  1,
		},
	})

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return log },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(log),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources with test data
			func() *sources.Sources { return mockSources },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus
			events.NewBus,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Initialize collector
			collector := colly.NewCollector()
			c.SetCollector(collector)

			// Test startup
			err := c.Start(context.Background(), "test")
			require.NoError(t, err)
		}),
	)

	require.NoError(t, app.Start(context.Background()))
	defer app.Stop(context.Background())
}

func TestCrawlerShutdown(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return log },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(log),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources
			func() *sources.Sources { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus
			events.NewBus,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Test shutdown
			err := c.Stop(context.Background())
			require.NoError(t, err)
		}),
	)

	require.NoError(t, app.Start(context.Background()))
	defer app.Stop(context.Background())
}

func TestSourceValidation(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return log },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(log),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources
			func() *sources.Sources { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus
			events.NewBus,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Test source validation
			err := c.Start(context.Background(), "invalid")
			require.Error(t, err)
		}),
	)

	require.NoError(t, app.Start(context.Background()))
	defer app.Stop(context.Background())
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return log },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(log),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources
			func() *sources.Sources { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus
			events.NewBus,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Test error handling
			err := c.Start(context.Background(), "error")
			require.Error(t, err)
		}),
	)

	require.NoError(t, app.Start(context.Background()))
	defer app.Stop(context.Background())
}

// writerWrapper implements io.Writer for the logger
type writerWrapper struct {
	logger common.Logger
}

// Write implements io.Writer interface
func (w *writerWrapper) Write(p []byte) (int, error) {
	w.logger.Debug(string(p))
	return len(p), nil
}

// NewDebugLogger creates a debug logger for testing.
func NewDebugLogger(logger common.Logger) io.Writer {
	return &writerWrapper{logger: logger}
}
