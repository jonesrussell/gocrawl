package collector_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
				Debugger:         &logger.CollyDebugger{Logger: logger.NewMockLogger()},
				Logger:           logger.NewMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source: &sources.Config{
					Selectors: struct {
						Article    string `yaml:"article"`
						Title      string `yaml:"title"`
						Date       string `yaml:"date"`
						Author     string `yaml:"author"`
						Categories string `yaml:"categories"`
					}{
						Article:    "article, .article",
						Title:      "h1",
						Date:       "time",
						Author:     ".author",
						Categories: "div.categories",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			params: collector.Params{
				BaseURL:          "",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: logger.NewMockLogger()},
				Logger:           logger.NewMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source: &sources.Config{
					Selectors: struct {
						Article    string `yaml:"article"`
						Title      string `yaml:"title"`
						Date       string `yaml:"date"`
						Author     string `yaml:"author"`
						Categories string `yaml:"categories"`
					}{
						Article:    "article, .article",
						Title:      "h1",
						Date:       "time",
						Author:     ".author",
						Categories: "div.categories",
					},
				},
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
				Debugger:         &logger.CollyDebugger{Logger: logger.NewMockLogger()},
				Logger:           logger.NewMockLogger(),
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          context.Background(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source: &sources.Config{
					Selectors: struct {
						Article    string `yaml:"article"`
						Title      string `yaml:"title"`
						Date       string `yaml:"date"`
						Author     string `yaml:"author"`
						Categories string `yaml:"categories"`
					}{
						Article:    "article, .article",
						Title:      "h1",
						Date:       "time",
						Author:     ".author",
						Categories: "div.categories",
					},
				},
			},
			wantErr:    true,
			wantErrMsg: "invalid base URL: not-a-url, must be a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations for the logger
			if !tt.wantErr {
				tt.params.Logger.(*logger.MockLogger).On("Debug", "Collector created", mock.Anything).Return()
			}

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

// TestCollectorCreation tests the collector creation with different URLs
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
			mockLogger := logger.NewMockLogger()
			// Set up mock expectations
			if !tt.wantErr {
				mockLogger.On("Debug", "Collector created", mock.Anything).Return()
			}

			params := collector.Params{
				BaseURL:   tt.baseURL,
				MaxDepth:  2,
				RateLimit: time.Second,
				Debugger: &logger.CollyDebugger{
					Logger: mockLogger,
				},
				Logger:           mockLogger,
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Context:          context.Background(),
				Source: &sources.Config{
					Selectors: struct {
						Article    string `yaml:"article"`
						Title      string `yaml:"title"`
						Date       string `yaml:"date"`
						Author     string `yaml:"author"`
						Categories string `yaml:"categories"`
					}{
						Article:    "article, .article",
						Title:      "h1",
						Date:       "time",
						Author:     ".author",
						Categories: "div.categories",
					},
				},
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
			ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
			Context:          context.Background(),
			Source: &sources.Config{
				Selectors: struct {
					Article    string `yaml:"article"`
					Title      string `yaml:"title"`
					Date       string `yaml:"date"`
					Author     string `yaml:"author"`
					Categories string `yaml:"categories"`
				}{
					Article:    "article, .article",
					Title:      "h1",
					Date:       "time",
					Author:     ".author",
					Categories: "div.categories",
				},
			},
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "logger is required", err.Error())
		require.Empty(t, result)
	})
}
