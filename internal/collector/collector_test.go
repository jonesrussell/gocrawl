package collector_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/jonesrussell/gocrawl/pkg/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// errMissingSource is copied from collector package for testing
const errMissingSource = "source configuration is required"

// createTestConfig creates a test config.Source with default selectors
func createTestConfig() *config.Source {
	return &config.Source{
		Name:         "test-source",
		URL:          "http://example.com",
		ArticleIndex: "test_articles",
		Index:        "test_content",
		RateLimit:    time.Second,
		MaxDepth:     2,
		Time:         []string{"03:00"},
		Selectors: config.SourceSelectors{
			Article: config.ArticleSelectors{
				Title:         "h1",
				Body:          ".article-body",
				PublishedTime: "time",
			},
		},
	}
}

// Helper function to create a mock logger
func newMockLogger() *testutils.MockLogger {
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	return mockLogger
}

// TestNew tests the New function of the collector package
func TestNew(t *testing.T) {
	mockLogger := newMockLogger()
	articleProcessor := article.NewService(mockLogger, config.DefaultArticleSelectors(), nil, "test-index")
	tests := []struct {
		name       string
		params     collector.Params
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid parameters",
			params: collector.Params{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   1 * time.Second,
				Debugger:    logger.NewCollyDebugger(mockLogger),
				Logger:      mockLogger,
				Parallelism: 2,
				RandomDelay: 2 * time.Second,
				Context:     t.Context(),
				ArticleProcessor: &article.ArticleProcessor{
					Logger:         mockLogger,
					ArticleService: articleProcessor,
					Storage:        nil,
					IndexName:      "test-index",
				},
				Source: createTestConfig(),
				Done:   make(chan struct{}),
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			params: collector.Params{
				BaseURL:     "",
				MaxDepth:    2,
				RateLimit:   1 * time.Second,
				Debugger:    logger.NewCollyDebugger(mockLogger),
				Logger:      mockLogger,
				Parallelism: 2,
				RandomDelay: 2 * time.Second,
				Context:     t.Context(),
				ArticleProcessor: &article.ArticleProcessor{
					Logger:         mockLogger,
					ArticleService: articleProcessor,
					Storage:        nil,
					IndexName:      "test-index",
				},
				Source: createTestConfig(),
				Done:   make(chan struct{}),
			},
			wantErr:    true,
			wantErrMsg: "base URL is required",
		},
		{
			name: "invalid base URL",
			params: collector.Params{
				BaseURL:     "not-a-url",
				MaxDepth:    2,
				RateLimit:   1 * time.Second,
				Debugger:    logger.NewCollyDebugger(mockLogger),
				Logger:      mockLogger,
				Parallelism: 2,
				RandomDelay: 2 * time.Second,
				Context:     t.Context(),
				ArticleProcessor: &article.ArticleProcessor{
					Logger:         mockLogger,
					ArticleService: articleProcessor,
					Storage:        nil,
					IndexName:      "test-index",
				},
				Source: createTestConfig(),
				Done:   make(chan struct{}),
			},
			wantErr:    true,
			wantErrMsg: "invalid base URL: not-a-url, must be a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations for the logger
			if !tt.wantErr {
				mockLogger := tt.params.Logger.(*testutils.MockLogger)
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
			}

			// Create the collector
			c, err := collector.New(tt.params)

			// Check error expectations
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, c)
		})
	}

	// Test for missing logger
	t.Run("missing logger", func(t *testing.T) {
		params := collector.Params{
			BaseURL: "http://example.com",
			Logger:  nil,
			ArticleProcessor: &article.ArticleProcessor{
				Logger:         mockLogger,
				ArticleService: articleProcessor,
				Storage:        nil,
				IndexName:      "test-index",
			},
			Context: t.Context(),
			Source:  createTestConfig(),
			Done:    make(chan struct{}),
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "logger is required", err.Error())
		require.Empty(t, result)
	})

	// Test for missing done channel
	t.Run("missing done channel", func(t *testing.T) {
		params := collector.Params{
			BaseURL: "http://example.com",
			Logger:  mockLogger,
			ArticleProcessor: &article.ArticleProcessor{
				Logger:         mockLogger,
				ArticleService: articleProcessor,
				Storage:        nil,
				IndexName:      "test-index",
			},
			Context: t.Context(),
			Source:  createTestConfig(),
			Done:    nil,
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "done channel is required", err.Error())
		require.Empty(t, result)
	})
}

// TestCollectorCreation tests the collector creation with different URLs
func TestCollectorCreation(t *testing.T) {
	mockLogger := newMockLogger()
	articleProcessor := article.NewService(mockLogger, config.DefaultArticleSelectors(), nil, "test-index")
	tests := []struct {
		name    string
		cfg     *config.Source
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  errMissingSource,
		},
		{
			name: "valid config",
			cfg: &config.Source{
				Name:         "test-source",
				URL:          "http://example.com",
				ArticleIndex: "test_articles",
				Index:        "test_content",
				RateLimit:    time.Second,
				MaxDepth:     2,
				Time:         []string{"03:00"},
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{
						Title:         "h1",
						Body:          ".article-body",
						PublishedTime: "time",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations for all debug calls
			if !tt.wantErr {
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
				mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
			}

			// Set base URL based on config
			baseURL := "http://example.com"
			if tt.cfg != nil {
				baseURL = tt.cfg.URL
			}

			params := collector.Params{
				BaseURL: baseURL,
				Logger:  mockLogger,
				ArticleProcessor: &article.ArticleProcessor{
					Logger:         mockLogger,
					ArticleService: articleProcessor,
					Storage:        nil,
					IndexName:      "test-index",
				},
				Context:     t.Context(),
				Source:      tt.cfg,
				Done:        make(chan struct{}),
				MaxDepth:    2,
				RateLimit:   time.Second,
				Parallelism: 2,
				RandomDelay: 2 * time.Second,
			}

			c, err := collector.New(params)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errMsg, err.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, c)
		})
	}
}
