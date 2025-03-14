package collector_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
)

// errMissingSource is copied from collector package for testing
const errMissingSource = "source configuration is required"

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

// Helper function to create a mock logger
func newMockLogger(t *testing.T) *logger.MockInterface {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	return mockLogger
}

// TestNew tests the New function of the collector package
func TestNew(t *testing.T) {
	mockLogger := newMockLogger(t)
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
				Debugger:         &logger.CollyDebugger{Logger: mockLogger},
				Logger:           mockLogger,
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: mockLogger},
				Source:           createTestConfig(),
				Done:             make(chan struct{}),
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			params: collector.Params{
				BaseURL:          "",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: mockLogger},
				Logger:           mockLogger,
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: mockLogger},
				Source:           createTestConfig(),
				Done:             make(chan struct{}),
			},
			wantErr:    true,
			wantErrMsg: "base URL is required",
		},
		{
			name: "invalid base URL",
			params: collector.Params{
				BaseURL:          "not-a-url",
				MaxDepth:         2,
				RateLimit:        1 * time.Second,
				Debugger:         &logger.CollyDebugger{Logger: mockLogger},
				Logger:           mockLogger,
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
				Context:          t.Context(),
				ArticleProcessor: &article.Processor{Logger: mockLogger},
				Source:           createTestConfig(),
				Done:             make(chan struct{}),
			},
			wantErr:    true,
			wantErrMsg: "invalid base URL: not-a-url, must be a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations for the logger
			if !tt.wantErr {
				mockLogger := tt.params.Logger.(*logger.MockInterface)
				mockLogger.EXPECT().Debug("Collector created",
					"baseURL", tt.params.BaseURL,
					"maxDepth", tt.params.MaxDepth,
					"rateLimit", tt.params.RateLimit,
					"parallelism", tt.params.Parallelism,
					"randomDelay", tt.params.RandomDelay,
				).AnyTimes()

				mockLogger.EXPECT().Debug("Collector configured",
					"allowedDomains", gomock.Any(),
					"disallowedDomains", gomock.Any(),
					"allowedURLs", gomock.Any(),
					"disallowedURLs", gomock.Any(),
				).AnyTimes()

				mockLogger.EXPECT().Debug("Collector callbacks configured").AnyTimes()
				mockLogger.EXPECT().Debug("Collector extensions configured").AnyTimes()
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
			BaseURL:          "http://example.com",
			Logger:           nil,
			ArticleProcessor: &article.Processor{Logger: mockLogger},
			Context:          t.Context(),
			Source:           createTestConfig(),
			Done:             make(chan struct{}),
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "logger is required", err.Error())
		require.Empty(t, result)
	})

	// Test for missing done channel
	t.Run("missing done channel", func(t *testing.T) {
		params := collector.Params{
			BaseURL:          "http://example.com",
			Logger:           mockLogger,
			ArticleProcessor: &article.Processor{Logger: mockLogger},
			Context:          t.Context(),
			Source:           createTestConfig(),
			Done:             nil,
		}

		result, err := collector.New(params)

		require.Error(t, err)
		require.Equal(t, "done channel is required", err.Error())
		require.Empty(t, result)
	})
}

// TestCollectorCreation tests the collector creation with different URLs
func TestCollectorCreation(t *testing.T) {
	mockLogger := newMockLogger(t)
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
				mockLogger.EXPECT().Debug("Collector created",
					"baseURL", gomock.Any(),
					"maxDepth", gomock.Any(),
					"rateLimit", gomock.Any(),
					"parallelism", gomock.Any(),
					"randomDelay", gomock.Any(),
				).AnyTimes()

				mockLogger.EXPECT().Debug("Collector configured",
					"allowedDomains", gomock.Any(),
					"disallowedDomains", gomock.Any(),
					"allowedURLs", gomock.Any(),
					"disallowedURLs", gomock.Any(),
				).AnyTimes()

				mockLogger.EXPECT().Debug("Collector callbacks configured").AnyTimes()
				mockLogger.EXPECT().Debug("Collector extensions configured").AnyTimes()
			}

			// Set base URL based on config
			baseURL := "http://example.com"
			if tt.cfg != nil {
				baseURL = tt.cfg.URL
			}

			params := collector.Params{
				BaseURL:          baseURL,
				Logger:           mockLogger,
				ArticleProcessor: &article.Processor{Logger: mockLogger},
				Context:          t.Context(),
				Source:           tt.cfg,
				Done:             make(chan struct{}),
				MaxDepth:         2,
				RateLimit:        time.Second,
				Parallelism:      2,
				RandomDelay:      2 * time.Second,
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
