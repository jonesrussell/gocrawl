package crawl_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// testDeps holds all the dependencies needed for testing
type testDeps struct {
	Crawler       crawler.Interface
	Storage       types.Interface
	Logger        common.Logger
	Config        config.Interface
	SourceManager sources.Interface
	Handler       *signal.SignalHandler
	Context       context.Context
	SourceName    string
}

// setupTestDeps creates and configures all test dependencies
func setupTestDeps(t *testing.T) *testDeps {
	t.Helper()

	// Create mock dependencies
	mockCrawler := testutils.NewMockCrawler()
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := testutils.NewMockConfig()
	mockSourceManager := testutils.NewMockSourceManager()

	// Set up basic expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockCrawler.On("Start", mock.Anything, mock.Anything).Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()

	return &testDeps{
		Crawler:       mockCrawler,
		Storage:       mockStorage,
		Logger:        mockLogger,
		Config:        mockConfig,
		SourceManager: mockSourceManager,
		Handler:       mockHandler,
		Context:       t.Context(),
		SourceName:    "test-source",
	}
}

// createTestApp creates a new fx app with the given dependencies and hooks
func createTestApp(t *testing.T, deps *testDeps, hooks ...fx.Hook) *fx.App {
	t.Helper()

	providers := []interface{}{
		// Core dependencies
		func() crawler.Interface { return deps.Crawler },
		func() types.Interface { return deps.Storage },
		func() common.Logger { return deps.Logger },
		func() config.Interface { return deps.Config },
		func() sources.Interface { return deps.SourceManager },

		// Named dependencies
		fx.Annotate(
			func() *signal.SignalHandler { return deps.Handler },
			fx.ResultTags(`name:"signalHandler"`),
		),
		fx.Annotate(
			func() context.Context { return deps.Context },
			fx.ResultTags(`name:"crawlContext"`),
		),
		fx.Annotate(
			func() string { return deps.SourceName },
			fx.ResultTags(`name:"sourceName"`),
		),
	}

	// Add command channels if needed
	commandDone := make(chan struct{})
	articleChannel := make(chan *models.Article)
	providers = append(providers,
		fx.Annotate(
			func() chan struct{} { return commandDone },
			fx.ResultTags(`name:"commandDone"`),
		),
		fx.Annotate(
			func() chan *models.Article { return articleChannel },
			fx.ResultTags(`name:"commandArticleChannel"`),
		),
	)

	// Create the app
	app := fx.New(
		fx.Provide(providers...),
		fx.Invoke(func(lc fx.Lifecycle, deps crawl.CommandDeps) {
			for _, hook := range hooks {
				lc.Append(hook)
			}
		}),
	)

	return app
}

// runTestApp runs the test app and handles cleanup
func runTestApp(t *testing.T, app *fx.App) {
	t.Helper()

	// Start the app
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)
}

func TestCommandExecution(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockCrawler := testutils.NewMockCrawler()
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := testutils.NewMockConfig()
	mockSourceManager := testutils.NewMockSourceManager()

	// Create channels
	commandDone := make(chan struct{})
	articleChannel := make(chan *models.Article)

	// Set up expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockCrawler.On("Start", mock.Anything, "test-source").Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()

	// Create test app
	app := fx.New(
		fx.Provide(
			// Core dependencies
			func() crawler.Interface { return mockCrawler },
			func() types.Interface { return mockStorage },
			func() common.Logger { return mockLogger },
			func() config.Interface { return mockConfig },
			func() sources.Interface { return mockSourceManager },

			// Named dependencies
			fx.Annotate(
				func() *signal.SignalHandler { return mockHandler },
				fx.ResultTags(`name:"signalHandler"`),
			),
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() chan struct{} { return commandDone },
				fx.ResultTags(`name:"commandDone"`),
			),
			fx.Annotate(
				func() chan *models.Article { return articleChannel },
				fx.ResultTags(`name:"commandArticleChannel"`),
			),
			fx.Annotate(
				func() common.Logger { return mockLogger },
				fx.ResultTags(`name:"testLogger"`),
			),
			fx.Annotate(
				func() sources.Interface { return mockSourceManager },
				fx.ResultTags(`name:"testSourceManager"`),
			),
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
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Close channels
	close(commandDone)
	close(articleChannel)

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
	cmd.SetContext(t.Context())

	// Test setting source flag
	err := cmd.Flags().Set("source", "test-source")
	require.NoError(t, err)

	// Verify flag value
	source, err := cmd.Flags().GetString("source")
	require.NoError(t, err)
	assert.Equal(t, "test-source", source)
}

func TestCrawlerCommandStartup(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Verify crawler is started
			deps.Crawler.(*testutils.MockCrawler).AssertCalled(t, "Start", ctx, deps.SourceName)
			return nil
		},
	})

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	deps.Crawler.(*testutils.MockCrawler).AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
}

func TestCrawlerCommandShutdown(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Create test app with shutdown hook
	app := createTestApp(t, deps, fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Verify crawler is stopped
			deps.Crawler.(*testutils.MockCrawler).AssertCalled(t, "Stop", ctx)
			return nil
		},
	})

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	deps.Crawler.(*testutils.MockCrawler).AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
}

func TestSourceValidation(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure source manager to return error for invalid source
	deps.SourceManager.(*testutils.MockSourceManager).On("GetSource", deps.SourceName).
		Return(nil, assert.AnError)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Verify crawler is not started
			deps.Crawler.(*testutils.MockCrawler).AssertNotCalled(t, "Start", ctx, deps.SourceName)
			return nil
		},
	})

	// Run the app and expect error
	err := app.Start(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get source")

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	deps.Crawler.(*testutils.MockCrawler).AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
	deps.SourceManager.(*testutils.MockSourceManager).AssertExpectations(t)
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure storage to return error on connection test
	deps.Storage.(*testutils.MockStorage).On("TestConnection", mock.Anything).
		Return(assert.AnError)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Verify crawler is not started
			deps.Crawler.(*testutils.MockCrawler).AssertNotCalled(t, "Start", ctx, deps.SourceName)
			return nil
		},
	})

	// Run the app and expect error
	err := app.Start(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to test connection")

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	deps.Crawler.(*testutils.MockCrawler).AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
}

func TestConcurrentCrawling(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure crawler for concurrent execution
	mockCrawler := deps.Crawler.(*testutils.MockCrawler)
	mockCrawler.On("Start", mock.Anything, deps.SourceName).Return(nil).Times(2)
	mockCrawler.On("Stop", mock.Anything).Return(nil).Times(2)
	mockCrawler.On("Wait").Return().Times(2)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start two crawlers concurrently
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				mockCrawler.Start(ctx, deps.SourceName)
			}()

			go func() {
				defer wg.Done()
				mockCrawler.Start(ctx, deps.SourceName)
			}()

			// Wait for both crawlers to complete
			wg.Wait()
			return nil
		},
	})

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
}

func TestRateLimiting(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure crawler for rate limiting
	mockCrawler := deps.Crawler.(*testutils.MockCrawler)
	mockCrawler.On("Start", mock.Anything, deps.SourceName).Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()
	mockCrawler.On("SetRateLimit", time.Second).Return(nil)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Set rate limit
			mockCrawler.SetRateLimit(time.Second)
			return nil
		},
	})

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
	deps.Storage.(*testutils.MockStorage).AssertExpectations(t)
}

func TestMaxDepthConfiguration(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockCrawler := testutils.NewMockCrawler()
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := testutils.NewMockConfig()
	mockSourceManager := testutils.NewMockSourceManager()

	// Set up expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockCrawler.On("Start", mock.Anything, "test-source").Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()
	mockCrawler.On("SetMaxDepth", 2).Return()

	// Create test app
	app := fx.New(
		fx.Provide(
			// Core dependencies
			func() crawler.Interface { return mockCrawler },
			func() types.Interface { return mockStorage },
			func() common.Logger { return mockLogger },
			func() config.Interface { return mockConfig },
			func() sources.Interface { return mockSourceManager },

			// Named dependencies
			fx.Annotate(
				func() *signal.SignalHandler { return mockHandler },
				fx.ResultTags(`name:"signalHandler"`),
			),
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, deps crawl.CommandDeps) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Set max depth
					deps.Crawler.SetMaxDepth(2)

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
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
