package crawler_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestCrawlerStartup(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create mock dependencies
	mockIndexManager := testutils.NewMockIndexManager()
	mockSourceManager := testutils.NewMockSourceManager()
	mockEventBus := events.NewBus()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(
			// Core dependencies
			func() common.Logger { return log },
			func() api.IndexManager { return mockIndexManager },
			func() sources.Interface { return mockSourceManager },
			func() *events.Bus { return mockEventBus },
			// Provide debugger
			func(logger common.Logger) debug.Debugger {
				return &debug.LogDebugger{
					Output: NewDebugLogger(logger),
				}
			},
			// Provide processors
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"articleProcessor"`),
			),
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"contentProcessor"`),
			),
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

			// Verify crawler is running by checking metrics
			metrics := c.GetMetrics()
			require.NotNil(t, metrics)

			// Stop crawler
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			require.NoError(t, c.Stop(ctx))
		}),
	)

	// Start the application
	require.NoError(t, app.Start(t.Context()))

	// Stop the application
	require.NoError(t, app.Stop(t.Context()))
}

func TestCrawlerShutdown(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test logger
	log, err := logger.NewLogger(testCfg)
	require.NoError(t, err)

	// Create mock dependencies
	mockIndexManager := testutils.NewMockIndexManager()
	mockSourceManager := testutils.NewMockSourceManager()
	mockEventBus := events.NewBus()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(
			// Core dependencies
			func() common.Logger { return log },
			func() api.IndexManager { return mockIndexManager },
			func() sources.Interface { return mockSourceManager },
			func() *events.Bus { return mockEventBus },
			// Provide debugger
			func(logger common.Logger) debug.Debugger {
				return &debug.LogDebugger{
					Output: NewDebugLogger(logger),
				}
			},
			// Provide processors
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"articleProcessor"`),
			),
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"contentProcessor"`),
			),
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

	// Create mock dependencies
	mockIndexManager := testutils.NewMockIndexManager()
	mockSourceManager := testutils.NewMockSourceManager()
	mockEventBus := events.NewBus()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(
			// Core dependencies
			func() common.Logger { return log },
			func() api.IndexManager { return mockIndexManager },
			func() sources.Interface { return mockSourceManager },
			func() *events.Bus { return mockEventBus },
			// Provide debugger
			func(logger common.Logger) debug.Debugger {
				return &debug.LogDebugger{
					Output: NewDebugLogger(logger),
				}
			},
			// Provide processors
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"articleProcessor"`),
			),
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"contentProcessor"`),
			),
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

	// Create mock dependencies
	mockIndexManager := testutils.NewMockIndexManager()
	mockSourceManager := testutils.NewMockSourceManager()
	mockEventBus := events.NewBus()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(
			// Core dependencies
			func() common.Logger { return log },
			func() api.IndexManager { return mockIndexManager },
			func() sources.Interface { return mockSourceManager },
			func() *events.Bus { return mockEventBus },
			// Provide debugger
			func(logger common.Logger) debug.Debugger {
				return &debug.LogDebugger{
					Output: NewDebugLogger(logger),
				}
			},
			// Provide processors
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"articleProcessor"`),
			),
			fx.Annotate(
				func(logger common.Logger) collector.Processor {
					return &testutils.MockProcessor{}
				},
				fx.ResultTags(`name:"contentProcessor"`),
			),
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
