package crawler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

// Process implements common.Processor
func (m *MockProcessor) Process(ctx context.Context, data any) error {
	return nil
}

// Start implements common.Processor
func (m *MockProcessor) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stop implements common.Processor
func (m *MockProcessor) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Ensure MockProcessor implements common.Processor
var _ common.Processor = (*MockProcessor)(nil)

func TestCrawlerStartup(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	testLogger, initErr := logger.NewLogger(testCfg)
	require.NoError(t, initErr)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1></body></html>"))
	}))
	defer server.Close()

	// Create mock sources with the test server URL
	mockSources := &sources.Sources{}
	mockSources.SetSources([]sources.Config{
		{
			Name:      "test",
			URL:       server.URL,
			RateLimit: time.Millisecond * 100, // Use a shorter rate limit for testing
			MaxDepth:  1,
		},
	})

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
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

			// Test startup with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			startErr := c.Start(ctx, "test")
			require.NoError(t, startErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

func TestCrawlerShutdown(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	testLogger, initErr := logger.NewLogger(testCfg)
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
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
			stopErr := c.Stop(t.Context())
			require.NoError(t, stopErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

func TestSourceValidation(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	testLogger, initErr := logger.NewLogger(testCfg)
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
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
			startErr := c.Start(t.Context(), "invalid")
			require.Error(t, startErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	testLogger, initErr := logger.NewLogger(testCfg)
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
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
			startErr := c.Start(t.Context(), "error")
			require.Error(t, startErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
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
