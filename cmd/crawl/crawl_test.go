package crawl_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/collector"
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
	Processors    []collector.Processor
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
		Processors:    []collector.Processor{},
	}
}

// createTestApp creates a new fx app with the given dependencies and hooks
func createTestApp(t *testing.T, deps *testDeps, hooks ...fx.Hook) *fx.App {
	t.Helper()

	providers := []any{
		// Core dependencies
		func() crawler.Interface { return deps.Crawler },
		func() types.Interface { return deps.Storage },
		func() common.Logger { return deps.Logger },
		func() config.Interface { return deps.Config },
		func() sources.Interface { return deps.SourceManager },
		func() *signal.SignalHandler { return deps.Handler },
		func() context.Context { return deps.Context },
		func() string { return deps.SourceName },
		func() []collector.Processor { return deps.Processors },
	}

	// Add command channels
	commandDone := make(chan struct{})
	articleChannel := make(chan *models.Article)
	providers = append(providers,
		func() chan struct{} { return commandDone },
		func() chan *models.Article { return articleChannel },
	)

	// Create the app
	app := fx.New(
		fx.Provide(providers...),
		fx.Invoke(func(lc fx.Lifecycle, deps crawl.CommandDeps) {
			// Add the provided hooks
			for _, hook := range hooks {
				lc.Append(hook)
			}

			// Add a default stop hook to clean up resources
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					// Close channels
					close(commandDone)
					close(articleChannel)
					return nil
				},
			})
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
			func() *signal.SignalHandler { return mockHandler },
			func() context.Context { return t.Context() },
			func() string { return "test-source" },
			func() chan struct{} { return commandDone },
			func() chan *models.Article { return articleChannel },
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

	// Configure crawler expectations
	mockCrawler := deps.Crawler.(*testutils.MockCrawler)
	mockCrawler.On("Start", mock.Anything, deps.SourceName).Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
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

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
}

func TestCrawlerCommandShutdown(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure crawler expectations
	mockCrawler := deps.Crawler.(*testutils.MockCrawler)
	mockCrawler.On("Start", mock.Anything, deps.SourceName).Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()

	// Create test app with startup and shutdown hooks
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
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

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
}

func TestSourceValidation(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure source manager to return error for invalid source
	mockSourceManager := deps.SourceManager.(*testutils.MockSourceManager)
	mockSourceManager.On("FindByName", deps.SourceName).
		Return(nil, assert.AnError)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Attempt to get source
			_, err := deps.SourceManager.FindByName(deps.SourceName)
			if err != nil {
				return err
			}
			return nil
		},
	})

	// Run the app and expect error
	err := app.Start(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "assert.AnError")

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	mockSourceManager.AssertExpectations(t)
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure storage to return error on connection test
	mockStorage := deps.Storage.(*testutils.MockStorage)
	mockStorage.On("TestConnection", mock.Anything).
		Return(assert.AnError)

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Test storage connection and return error
			return deps.Storage.TestConnection(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	// Run the app and expect error
	err := app.Start(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "assert.AnError")

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	mockStorage.AssertExpectations(t)
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
			if err := deps.Crawler.SetRateLimit(time.Second); err != nil {
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

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
}

func TestMaxDepthConfiguration(t *testing.T) {
	t.Parallel()

	// Set up test dependencies
	deps := setupTestDeps(t)

	// Configure crawler for max depth
	mockCrawler := deps.Crawler.(*testutils.MockCrawler)
	mockCrawler.On("Start", mock.Anything, deps.SourceName).Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil)
	mockCrawler.On("Wait").Return()
	mockCrawler.On("SetMaxDepth", 2).Return()

	// Create test app with startup hook
	app := createTestApp(t, deps, fx.Hook{
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

	// Run the app
	runTestApp(t, app)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
}
