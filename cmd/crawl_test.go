package cmd_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// setupTestConfig creates a temporary sources.yml file for testing
func setupTestConfig(t *testing.T) {
	t.Helper()

	// Create a temporary directory and change to it
	t.Chdir(t.TempDir())

	// Create test sources.yml
	content := []byte(`
sources:
  - name: test-source
    url: https://example.com
    index: test_content
    article_index: test_articles
    rate_limit: 1s
    max_depth: 2
`)
	require.NoError(t, os.WriteFile("sources.yml", content, 0644))
}

func TestStartCrawl(t *testing.T) {
	setupTestConfig(t)

	tests := []struct {
		name    string
		setup   func(*testing.T) (*fxtest.App, cmd.CrawlParams)
		wantErr bool
	}{
		{
			name: "successful crawl",
			setup: func(t *testing.T) (*fxtest.App, cmd.CrawlParams) {
				mockCrawler := crawler.NewMockCrawler()
				mockLogger := logger.NewMockLogger()
				mockStorage := storage.NewMockStorage()
				cmd.SetSourceName("test-source") // Using exported function to set source name

				sources, err := sources.Load("sources.yml")
				require.NoError(t, err)
				sources.Logger = mockLogger

				doneChan := make(chan struct{})

				mockCrawler.On("SetCollector", mock.Anything).Return()
				mockCrawler.On("Start", mock.Anything, mock.Anything).Return(nil)
				mockCrawler.On("Stop").Return()

				var lifecycle fx.Lifecycle
				app := fxtest.New(t,
					fx.Populate(&lifecycle),
				)

				params := cmd.CrawlParams{
					Lifecycle:       lifecycle,
					Sources:         sources,
					CrawlerInstance: mockCrawler,
					Logger:          mockLogger,
					Done:            doneChan,
					Processors: []models.ContentProcessor{
						&article.Processor{
							Logger:    mockLogger,
							Storage:   mockStorage,
							IndexName: "test_articles",
						},
						&article.Processor{
							Logger:    mockLogger,
							Storage:   mockStorage,
							IndexName: "test_content",
						},
					},
				}
				return app, params
			},
			wantErr: false,
		},
		{
			name: "nil crawler instance",
			setup: func(t *testing.T) (*fxtest.App, cmd.CrawlParams) {
				mockLogger := logger.NewMockLogger()
				sources, err := sources.Load("sources.yml")
				require.NoError(t, err)
				sources.Logger = mockLogger

				var lifecycle fx.Lifecycle
				app := fxtest.New(t,
					fx.Populate(&lifecycle),
				)

				params := cmd.CrawlParams{
					Lifecycle:       lifecycle,
					Sources:         sources,
					CrawlerInstance: nil,
					Logger:          mockLogger,
					Done:            make(chan struct{}),
				}
				return app, params
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, params := tt.setup(t)
			err := cmd.StartCrawl(params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Start the application
			app.RequireStart()

			// Give the goroutine time to start
			time.Sleep(100 * time.Millisecond)

			// Stop the application
			app.RequireStop()
		})
	}
}

func TestCrawlCommand(t *testing.T) {
	setupTestConfig(t)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no source name provided",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"source1", "source2"},
			wantErr: true,
		},
		{
			name:    "valid source name",
			args:    []string{"test-source"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test app with mocked dependencies
			var testApp *fxtest.App
			var err error

			if !tt.wantErr {
				mockLogger := logger.NewMockLogger()
				mockStorage := storage.NewMockStorage()
				mockCrawler := crawler.NewMockCrawler()

				mockCrawler.On("SetCollector", mock.Anything).Return()
				mockCrawler.On("Start", mock.Anything, mock.Anything).Return(nil)
				mockCrawler.On("Stop").Return()

				testApp = fxtest.New(t,
					fx.Supply(mockLogger),
					fx.Supply(mockStorage),
					fx.Supply(mockCrawler),
					fx.Provide(
						func() crawler.Interface { return mockCrawler },
					),
				)
			}

			// Test command argument validation
			err = cmd.CrawlCmd.Args(cmd.CrawlCmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, testApp)

			// Test command execution
			if testApp != nil {
				testApp.RequireStart()
				defer testApp.RequireStop()
			}
		})
	}
}

func TestCrawlCommand_Args(t *testing.T) {
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
			name:    "too many arguments",
			args:    []string{"source1", "source2"},
			wantErr: true,
		},
		{
			name:    "valid argument",
			args:    []string{"test-source"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.CrawlCmd.Args(cmd.CrawlCmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCrawlCommand_Execute(t *testing.T) {
	setupTestConfig(t)

	// Set up mocks
	mockLogger := logger.NewMockLogger()
	mockStorage := storage.NewMockStorage()
	mockCrawler := crawler.NewMockCrawler()

	mockCrawler.On("SetCollector", mock.Anything).Return()
	mockCrawler.On("Start", mock.Anything, mock.Anything).Return(nil)
	mockCrawler.On("Stop").Return()

	// Create a test app
	app := fxtest.New(t,
		fx.Supply(mockLogger),
		fx.Supply(mockStorage),
		fx.Supply(mockCrawler),
		fx.Provide(
			func() crawler.Interface { return mockCrawler },
		),
		common.Module,
		crawler.Module,
		article.Module,
		content.Module,
	)

	// Start the test app
	app.RequireStart()
	defer app.RequireStop()

	// Test execution
	cmd := cmd.CrawlCmd
	cmd.SetArgs([]string{"test-source"})

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	// Execute the command
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	// Verify mock expectations
	mockCrawler.AssertExpectations(t)
}

func TestCrawlCommand_ExecuteError(t *testing.T) {
	setupTestConfig(t)

	// Test with invalid source name
	cmd := cmd.CrawlCmd
	cmd.SetArgs([]string{"nonexistent-source"})

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	// Execute the command
	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
}
