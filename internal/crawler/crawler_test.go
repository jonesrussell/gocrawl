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
	"github.com/jonesrussell/gocrawl/internal/config"
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

	// Create test app with all required dependencies
	app := fx.New(
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
			// Provide event bus
			events.NewBus,
			// Provide article processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide crawler
			crawler.ProvideCrawler,
		),
		fx.Invoke(
			func(c crawler.Interface) {
				// Test crawler startup
				ctx := t.Context()
				startErr := c.Start(ctx, "test-source")
				require.NoError(t, startErr)
				defer c.Stop(ctx)
			},
		),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
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
		fx.Supply(testCfg),
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
			// Provide event bus
			events.NewBus,
			// Provide article processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"shutdownArticleProcessor"`),
			),
			// Provide content processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"shutdownContentProcessor"`),
			),
			// Provide crawler
			crawler.ProvideCrawler,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Start crawler in background
			go func() {
				if startErr := c.Start(t.Context(), "test_source"); startErr != nil {
					t.Errorf("Failed to start crawler: %v", startErr)
				}
			}()

			// Wait for crawler to start
			time.Sleep(100 * time.Millisecond)

			// Start a long-running crawl
			done := make(chan bool)
			go func() {
				// Simulate a long-running crawl
				time.Sleep(2 * time.Second)
				done <- true
			}()

			// Wait for crawl to start
			time.Sleep(50 * time.Millisecond)

			// Stop crawler with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			require.NoError(t, c.Stop(ctx))

			// Wait for crawl to complete
			select {
			case <-done:
				// Crawl completed successfully
			case <-time.After(3 * time.Second):
				t.Fatal("Crawl did not complete within timeout")
			}
		}),
	)

	// Start the application
	require.NoError(t, app.Start(t.Context()))

	// Stop the application
	require.NoError(t, app.Stop(t.Context()))
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
		fx.Supply(testCfg),
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
			// Provide event bus
			events.NewBus,
			// Provide article processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"validationArticleProcessor"`),
			),
			// Provide content processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"validationContentProcessor"`),
			),
			// Provide crawler
			crawler.ProvideCrawler,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Test valid source
			source := &config.Source{
				Name:      "test_source",
				URL:       "https://example.com",
				MaxDepth:  1,
				RateLimit: time.Second,
			}
			require.NoError(t, source.Validate())

			// Test invalid source (missing URL)
			invalidSource := &config.Source{
				Name:      "invalid_source",
				MaxDepth:  1,
				RateLimit: time.Second,
			}
			require.Error(t, invalidSource.Validate())

			// Test invalid source (missing name)
			invalidSource = &config.Source{
				URL:       "https://example.com",
				MaxDepth:  1,
				RateLimit: time.Second,
			}
			require.Error(t, invalidSource.Validate())
		}),
	)

	// Start the application
	require.NoError(t, app.Start(t.Context()))

	// Stop the application
	require.NoError(t, app.Stop(t.Context()))
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Create test configuration with invalid settings
	testCfg := testutils.NewMockConfig()
	testCfg.Crawler.MaxDepth = -1 // Invalid value

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create test app with all required dependencies
	app := fx.New(
		fx.Supply(testCfg),
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
			// Provide event bus
			events.NewBus,
			// Provide article processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"errorArticleProcessor"`),
			),
			// Provide content processor
			fx.Annotate(
				func() *MockProcessor {
					return &MockProcessor{}
				},
				fx.As(new(common.Processor)),
				fx.ResultTags(`name:"errorContentProcessor"`),
			),
			// Provide crawler
			crawler.ProvideCrawler,
		),
		fx.Invoke(func(c crawler.Interface) {
			// Attempt to start crawler with invalid config
			startErr := c.Start(t.Context(), "test_source")
			require.Error(t, startErr)

			// Test error handling during crawl
			source := &config.Source{
				Name:      "invalid_source",
				URL:       "https://invalid-url.com",
				MaxDepth:  1,
				RateLimit: time.Second,
			}
			validateErr := source.Validate()
			require.NoError(t, validateErr)
		}),
	)

	// Start the application
	require.NoError(t, app.Start(t.Context()))

	// Stop the application
	require.NoError(t, app.Stop(t.Context()))
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
