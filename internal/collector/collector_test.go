package collector_test

import (
	"context"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a mock logger
func newMockLogger() *logger.MockLogger {
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return() // Set up expectation for Debug call
	mockLogger.On("Error", mock.Anything, mock.Anything).Return() // Set up expectation for Error call
	return mockLogger
}

// TestNew tests the New function of the collector package
func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		params     collector.Params
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid parameters",
			params: collector.Params{
				BaseURL:          "http://example.com",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: newMockLogger()},
				Logger:           newMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: newMockLogger()},
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			params: collector.Params{
				BaseURL:          "",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: newMockLogger()},
				Logger:           newMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: newMockLogger()},
			},
			wantErr:    true,
			wantErrMsg: "base URL cannot be empty",
		},
		{
			name: "invalid base URL",
			params: collector.Params{
				BaseURL:          "not-a-url",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: newMockLogger()},
				Logger:           newMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: newMockLogger()},
			},
			wantErr:    true,
			wantErrMsg: "invalid base URL: not-a-url, must be a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := collector.New(tt.params)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantErrMsg, err.Error())
				require.Empty(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result.Collector)
			}
		})
	}
}

// TestConfigureLogging tests the ConfigureLogging function
func TestConfigureLogging(t *testing.T) {
	c := colly.NewCollector()
	testLogger := newMockLogger() // Use the helper function

	// Set expectation for the "Error occurred" log message
	testLogger.On("Error", "Error occurred", mock.Anything).Return()

	// Call the ConfigureLogging function
	collector.ConfigureLogging(c, testLogger)

	// Set expectation for the "Requesting URL" log message
	testLogger.On("Debug", "Requesting URL", mock.Anything).Return()

	// Simulate a request to test logging
	c.OnRequest(func(r *colly.Request) {
		testLogger.Debug("Requesting URL", r.URL.String())
	})

	// Trigger a request to see if logging works
	c.Visit("http://example.com")

	// Check if the logger received the expected log messages
	messages := testLogger.GetMessages()
	require.Contains(t, messages, "Requesting URL")
}

func TestCollectorCreation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid URL",
			baseURL: "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty URL",
			baseURL: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := newMockLogger()

			params := collector.Params{
				BaseURL:   tt.baseURL,
				MaxDepth:  2,
				RateLimit: time.Second,
				Debugger: &logger.CollyDebugger{
					Logger: mockLogger,
				},
				Logger:           mockLogger,
				ArticleProcessor: &article.Processor{Logger: newMockLogger()},
				Context:          context.Background(),
			}

			result, err := collector.New(params)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result.Collector)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result.Collector)
			}
		})
	}

	// Test for missing logger
	t.Run("missing logger", func(t *testing.T) {
		params := collector.Params{
			BaseURL:          "http://example.com",
			Logger:           nil,
			ArticleProcessor: &article.Processor{Logger: newMockLogger()},
			Context:          context.Background(),
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "article processor is required", err.Error())
		require.Empty(t, result)
	})
}

// TestNewCollector tests the New function of the collector package
func TestNewCollector(t *testing.T) {
	mockLogger := logger.NewMockLogger()

	params := collector.Params{
		BaseURL:          "http://example.com",
		MaxDepth:         3,
		RateLimit:        time.Second,
		Debugger:         logger.NewCollyDebugger(mockLogger),
		Logger:           mockLogger,
		ArticleProcessor: &article.Processor{Logger: newMockLogger()},
		Context:          context.Background(), // Initialize context
	}

	// Set expectation for the "Collector created" log message
	mockLogger.On("Debug", "Collector created", mock.Anything).Return()

	// Create the collector using the collector module
	collectorResult, err := collector.New(params) // Pass the mockCrawler here

	require.NoError(t, err)
	require.NotNil(t, collectorResult.Collector)
	// Add additional assertions as needed
}

// TestNewCollector_MissingLogger tests the New function of the collector package
func TestNewCollector_MissingLogger(t *testing.T) {
	params := collector.Params{
		BaseURL:          "http://example.com", // Provide a valid base URL
		Logger:           nil,                  // Pass nil for the logger
		ArticleProcessor: &article.Processor{Logger: newMockLogger()},
		Context:          context.Background(), // Initialize context
	}

	result, err := collector.New(params) // Pass nil for the crawler

	require.Error(t, err)
	require.Equal(t, "crawler instance is required", err.Error())
	require.Empty(t, result)
}
