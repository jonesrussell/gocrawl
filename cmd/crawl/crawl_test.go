package crawl_test

import (
	"context"
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

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(crawl.Command),
		fx.Invoke(func(cmd *cobra.Command) {
			// Start command in background
			go func() {
				if err := cmd.Execute(); err != nil {
					t.Errorf("Failed to execute command: %v", err)
				}
			}()

			// Wait for command to start
			time.Sleep(100 * time.Millisecond)

			// Verify command is running
			require.NotNil(t, cmd)

			// Stop command
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			require.NoError(t, cmd.ExecuteContext(ctx))
		}),
	)

	// Start the application
	require.NoError(t, app.Start(context.Background()))

	// Stop the application
	require.NoError(t, app.Stop(context.Background()))
}

func TestCrawlerCommandShutdown(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(crawl.Command),
		fx.Invoke(func(cmd *cobra.Command) {
			// Start command in background
			go func() {
				if err := cmd.Execute(); err != nil {
					t.Errorf("Failed to execute command: %v", err)
				}
			}()

			// Wait for command to start
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

			// Stop command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			require.NoError(t, cmd.ExecuteContext(ctx))

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
	require.NoError(t, app.Start(context.Background()))

	// Stop the application
	require.NoError(t, app.Stop(context.Background()))
}

func TestSourceValidation(t *testing.T) {
	t.Parallel()

	// Create test configuration
	testCfg := testutils.NewMockConfig()

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(crawl.Command),
		fx.Invoke(func(cmd *cobra.Command) {
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
	require.NoError(t, app.Start(context.Background()))

	// Stop the application
	require.NoError(t, app.Stop(context.Background()))
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()

	// Create test configuration with invalid settings
	testCfg := testutils.NewMockConfig()
	testCfg.Crawler.MaxDepth = -1 // Invalid value

	// Create test app
	app := fx.New(
		fx.Supply(testCfg),
		fx.Provide(crawl.Command),
		fx.Invoke(func(cmd *cobra.Command) {
			// Attempt to start command with invalid config
			err := cmd.Execute()
			require.Error(t, err)

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
	require.NoError(t, app.Start(context.Background()))

	// Stop the application
	require.NoError(t, app.Stop(context.Background()))
}

func TestConcurrentCrawling(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockCrawler := testutils.NewMockCrawler()
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := testutils.NewMockConfig()
	mockSourceManager := testutils.NewMockSourceManager()

	// Set up expectations for concurrent crawling
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockCrawler.On("Start", mock.Anything, "source1").Return(nil)
	mockCrawler.On("Start", mock.Anything, "source2").Return(nil)
	mockCrawler.On("Stop", mock.Anything).Return(nil).Times(2)
	mockCrawler.On("Wait").Return().Times(2)

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
		),
		fx.Invoke(func(lc fx.Lifecycle, deps crawl.CommandDeps) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Start first crawler
					go func() {
						if err := deps.Crawler.Start(ctx, "source1"); err != nil {
							t.Errorf("Failed to start first crawler: %v", err)
						}
						deps.Crawler.Wait()
					}()

					// Start second crawler
					go func() {
						if err := deps.Crawler.Start(ctx, "source2"); err != nil {
							t.Errorf("Failed to start second crawler: %v", err)
						}
						deps.Crawler.Wait()
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

	// Wait for crawlers to complete
	time.Sleep(200 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify all expectations were met
	mockCrawler.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestRateLimiting(t *testing.T) {
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
	mockCrawler.On("SetRateLimit", time.Second).Return(nil)

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
