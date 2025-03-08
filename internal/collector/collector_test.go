package collector_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
)

// ArticleSelectors represents the article selectors structure
type ArticleSelectors struct {
	Container     string `yaml:"container,omitempty"`
	Title         string `yaml:"title"`
	Body          string `yaml:"body"`
	Intro         string `yaml:"intro,omitempty"`
	Byline        string `yaml:"byline,omitempty"`
	PublishedTime string `yaml:"published_time"`
	TimeAgo       string `yaml:"time_ago,omitempty"`
	JSONLD        string `yaml:"json_ld,omitempty"`
	Section       string `yaml:"section,omitempty"`
	Keywords      string `yaml:"keywords,omitempty"`
	Description   string `yaml:"description,omitempty"`
	OgTitle       string `yaml:"og_title,omitempty"`
	OgDescription string `yaml:"og_description,omitempty"`
	OgImage       string `yaml:"og_image,omitempty"`
	OgURL         string `yaml:"og_url,omitempty"`
	Canonical     string `yaml:"canonical,omitempty"`
	WordCount     string `yaml:"word_count,omitempty"`
	PublishDate   string `yaml:"publish_date,omitempty"`
	Category      string `yaml:"category,omitempty"`
	Tags          string `yaml:"tags,omitempty"`
	Author        string `yaml:"author,omitempty"`
	BylineName    string `yaml:"byline_name,omitempty"`
}

// Selectors represents the selectors structure
type Selectors struct {
	Article ArticleSelectors `yaml:"article"`
}

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
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source:           createTestConfig(),
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
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source:           createTestConfig(),
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
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Source:           createTestConfig(),
			},
			wantErr:    true,
			wantErrMsg: "invalid base URL: not-a-url, must be a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations for the logger
			if !tt.wantErr {
				mockLogger := tt.params.Logger.(*logger.MockLogger)
				// Set up expectations for all debug calls
				mockLogger.On("Debug", "Collector created",
					"baseURL", tt.params.BaseURL,
					"maxDepth", tt.params.MaxDepth,
					"rateLimit", tt.params.RateLimit,
					"parallelism", tt.params.Parallelism,
				).Return()
				mockLogger.On("Debug", "Setting up article processing", "tag", "collector/content").Return()
				mockLogger.On("Debug", "Setting up HTML processing", "tag", "collector/content").Return()
				mockLogger.On("Debug", "Setting up link following", "tag", "collector/content").Return()
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
			// Set up mock expectations for all debug calls
			if !tt.wantErr {
				mockLogger.On("Debug", "Collector created",
					"baseURL", tt.baseURL,
					"maxDepth", 2,
					"rateLimit", time.Second,
					"parallelism", 2,
				).Return()
				mockLogger.On("Debug", "Setting up article processing", "tag", "collector/content").Return()
				mockLogger.On("Debug", "Setting up HTML processing", "tag", "collector/content").Return()
				mockLogger.On("Debug", "Setting up link following", "tag", "collector/content").Return()
			}

			// Create test config
			cfg := createTestConfig()

			params := collector.Params{
				BaseURL:     tt.baseURL,
				MaxDepth:    2,
				RateLimit:   time.Second,
				Parallelism: 2,
				Debugger: &logger.CollyDebugger{
					Logger: mockLogger,
				},
				Logger:           mockLogger,
				ArticleProcessor: &article.Processor{Logger: logger.NewMockLogger()},
				Context:          t.Context(),
				Source:           cfg,
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
			Context:          t.Context(),
			Source:           createTestConfig(),
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "logger is required", err.Error())
		require.Empty(t, result)
	})
}
