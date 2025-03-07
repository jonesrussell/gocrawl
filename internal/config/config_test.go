package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestNew(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")     // Name of the file without extension
	viper.AddConfigPath("./testdata") // Path to the testdata directory

	err := viper.ReadInConfig()
	require.NoError(t, err)

	cfg, err := config.New()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test", cfg.App.Environment)
	require.Equal(t, "debug", cfg.Log.Level)
	require.True(t, cfg.Log.Debug)
	require.Equal(t, "http://test.example.com", cfg.Crawler.BaseURL)
	require.Equal(t, 5, cfg.Crawler.MaxDepth)
	require.Equal(t, 2*time.Second, cfg.Crawler.RateLimit)
	require.Equal(t, "http://localhost:9200", cfg.Elasticsearch.URL)
	require.Equal(t, "test_user", cfg.Elasticsearch.Username)
	require.Equal(t, "test_pass", cfg.Elasticsearch.Password)
	require.Equal(t, "test_apikey", cfg.Elasticsearch.APIKey)
	require.Equal(t, "test_index", cfg.Elasticsearch.IndexName)
	require.False(t, cfg.Elasticsearch.SkipTLS)
}

func TestParseRateLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "valid duration 1s",
			input:    "1s",
			expected: time.Second,
			wantErr:  false,
		},
		{
			name:     "valid duration 2m",
			input:    "2m",
			expected: 2 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Second,
			wantErr:  true,
		},
		{
			name:     "invalid duration",
			input:    "invalid",
			expected: time.Second,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := config.ParseRateLimit(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, duration)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr error
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					URL: "http://localhost:9200",
				},
			},
			wantErr: nil,
		},
		{
			name: "missing elasticsearch URL",
			cfg: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					URL: "",
				},
			},
			wantErr: config.ErrMissingElasticURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfig(tt.cfg)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid config file",
			path:    "testdata/config.yml",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "testdata/nonexistent.yml",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			path:    "testdata/invalid.yml",
			wantErr: true,
		},
	}

	// Create invalid YAML file for testing
	invalidYAML := []byte("invalid: yaml: content")
	err := os.WriteFile("testdata/invalid.yml", invalidYAML, 0644)
	require.NoError(t, err)
	defer os.Remove("testdata/invalid.yml")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, loadErr := config.LoadConfig(tt.path)
			if tt.wantErr {
				require.Error(t, loadErr)
			} else {
				require.NoError(t, loadErr)
			}
		})
	}
}

func TestInitializeConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfgFile string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "with config file",
			cfgFile: "testdata/config.yml",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
				"APP_ENV":   "test",
			},
			wantErr: false,
		},
		{
			name:    "without config file",
			cfgFile: "",
			envVars: map[string]string{
				"LOG_LEVEL": "info",
				"APP_ENV":   "development",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := config.InitializeConfig(tt.cfgFile)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				require.Equal(t, tt.envVars["LOG_LEVEL"], cfg.Log.Level)
				require.Equal(t, tt.envVars["APP_ENV"], cfg.App.Environment)
			}
		})
	}
}

func TestCrawlerConfig_Setters(t *testing.T) {
	cfg := &config.CrawlerConfig{}

	// Test SetMaxDepth
	cfg.SetMaxDepth(5)
	assert.Equal(t, 5, cfg.MaxDepth)
	assert.Equal(t, 5, viper.GetInt(config.CrawlerMaxDepthKey))

	// Test SetRateLimit
	rateLimit := 2 * time.Second
	cfg.SetRateLimit(rateLimit)
	assert.Equal(t, rateLimit, cfg.RateLimit)
	assert.Equal(t, rateLimit.String(), viper.GetString(config.CrawlerRateLimitKey))

	// Test SetBaseURL
	baseURL := "http://example.com"
	cfg.SetBaseURL(baseURL)
	assert.Equal(t, baseURL, cfg.BaseURL)
	assert.Equal(t, baseURL, viper.GetString(config.CrawlerBaseURLKey))

	// Test SetIndexName
	indexName := "test_index"
	cfg.SetIndexName(indexName)
	assert.Equal(t, indexName, cfg.IndexName)
	assert.Equal(t, indexName, viper.GetString(config.ElasticIndexNameKey))
}

func TestDefaultArticleSelectors(t *testing.T) {
	selectors := config.DefaultArticleSelectors()

	// Test that default selectors are not empty
	assert.NotEmpty(t, selectors.Container)
	assert.NotEmpty(t, selectors.Title)
	assert.NotEmpty(t, selectors.Body)
	assert.NotEmpty(t, selectors.Intro)
	assert.NotEmpty(t, selectors.Byline)
	assert.NotEmpty(t, selectors.PublishedTime)
	assert.NotEmpty(t, selectors.TimeAgo)
	assert.NotEmpty(t, selectors.JSONLD)
	assert.NotEmpty(t, selectors.Section)
	assert.NotEmpty(t, selectors.Keywords)
	assert.NotEmpty(t, selectors.Description)
	assert.NotEmpty(t, selectors.OGTitle)
	assert.NotEmpty(t, selectors.OGDescription)
	assert.NotEmpty(t, selectors.OGImage)
	assert.NotEmpty(t, selectors.OgURL)
	assert.NotEmpty(t, selectors.Canonical)
}

func TestNewHTTPTransport(t *testing.T) {
	transport := config.NewHTTPTransport()
	assert.NotNil(t, transport)
}
