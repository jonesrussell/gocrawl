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
	tests := []struct {
		name     string
		setup    func()
		validate func(*testing.T, *config.Config, error)
	}{
		{
			name: "valid configuration",
			setup: func() {
				viper.SetConfigType("yaml")
				viper.SetConfigName("config")
				viper.AddConfigPath("./testdata")
				require.NoError(t, viper.ReadInConfig())
			},
			validate: func(t *testing.T, cfg *config.Config, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				require.Equal(t, "test", cfg.App.Environment)
				require.Equal(t, "debug", cfg.Log.Level)
				require.True(t, cfg.Log.Debug)
				require.Equal(t, "http://test.example.com", cfg.Crawler.BaseURL)
				require.Equal(t, 5, cfg.Crawler.MaxDepth)
				require.Equal(t, 2*time.Second, cfg.Crawler.RateLimit)
				require.Equal(t, 2, cfg.Crawler.Parallelism)
				require.Equal(t, []string{"http://localhost:9200"}, cfg.Elasticsearch.Addresses)
				require.Equal(t, "test_user", cfg.Elasticsearch.Username)
				require.Equal(t, "test_pass", cfg.Elasticsearch.Password)
				require.Equal(t, "test_apikey", cfg.Elasticsearch.APIKey)
				require.Equal(t, "test_index", cfg.Elasticsearch.IndexName)
				require.True(t, cfg.Elasticsearch.TLS.SkipVerify)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg, err := config.New()
			tt.validate(t, cfg, err)
		})
	}
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
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					Parallelism: 2,
					MaxDepth:    2,
					RateLimit:   time.Second,
					RandomDelay: time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "missing elastic addresses",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid parallelism",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					Parallelism: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "negative max depth",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					MaxDepth: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					RateLimit: -time.Second,
				},
			},
			wantErr: true,
		},
		{
			name: "negative random delay",
			config: &config.Config{
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					RandomDelay: -time.Second,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
				"log.level":           "debug",
				"app.environment":     "test",
				"crawler.parallelism": "2",
			},
			wantErr: false,
		},
		{
			name:    "without config file",
			cfgFile: "",
			envVars: map[string]string{
				"log.level":           "info",
				"app.environment":     "development",
				"crawler.parallelism": "1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Viper configuration
			viper.Reset()

			// Set config file if provided
			if tt.cfgFile != "" {
				viper.SetConfigFile(tt.cfgFile)
			}

			// Set environment variables
			for k, v := range tt.envVars {
				viper.Set(k, v)
			}

			// Initialize config
			cfg, err := config.InitializeConfig()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Verify environment variables were properly set
			for k, v := range tt.envVars {
				switch k {
				case "log.level":
					require.Equal(t, v, cfg.Log.Level)
				case "app.environment":
					require.Equal(t, v, cfg.App.Environment)
				case "crawler.parallelism":
					parallelism := viper.GetInt("crawler.parallelism")
					require.Equal(t, parallelism, cfg.Crawler.Parallelism)
				}
			}
		})
	}
}

func TestSetMaxDepth(t *testing.T) {
	cfg := &config.CrawlerConfig{}
	cfg.SetMaxDepth(5)
	require.Equal(t, 5, cfg.MaxDepth)
	require.Equal(t, 5, viper.GetInt("crawler.max_depth"))
}

func TestSetRateLimit(t *testing.T) {
	cfg := &config.CrawlerConfig{}
	rateLimit := 2 * time.Second
	cfg.SetRateLimit(rateLimit)
	require.Equal(t, rateLimit, cfg.RateLimit)
	require.Equal(t, "2s", viper.GetString("crawler.rate_limit"))
}

func TestSetBaseURL(t *testing.T) {
	cfg := &config.CrawlerConfig{}
	url := "http://example.com"
	cfg.SetBaseURL(url)
	require.Equal(t, url, cfg.BaseURL)
	require.Equal(t, url, viper.GetString("crawler.base_url"))
}

func TestSetIndexName(t *testing.T) {
	cfg := &config.CrawlerConfig{}
	index := "test_index"
	cfg.SetIndexName(index)
	require.Equal(t, index, cfg.IndexName)
	require.Equal(t, index, viper.GetString("elasticsearch.index_name"))
}

func TestCrawlerConfig_Setters(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*config.CrawlerConfig)
		validate func(*testing.T, *config.CrawlerConfig)
	}{
		{
			name: "set max depth",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetMaxDepth(5)
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, 5, cfg.MaxDepth)
				require.Equal(t, 5, viper.GetInt("crawler.max_depth"))
			},
		},
		{
			name: "set rate limit",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetRateLimit(2 * time.Second)
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, 2*time.Second, cfg.RateLimit)
				require.Equal(t, "2s", viper.GetString("crawler.rate_limit"))
			},
		},
		{
			name: "set base url",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetBaseURL("http://example.com")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "http://example.com", cfg.BaseURL)
				require.Equal(t, "http://example.com", viper.GetString("crawler.base_url"))
			},
		},
		{
			name: "set index name",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetIndexName("test_index")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "test_index", cfg.IndexName)
				require.Equal(t, "test_index", viper.GetString("elasticsearch.index_name"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			cfg := &config.CrawlerConfig{}
			tt.setup(cfg)
			tt.validate(t, cfg)
		})
	}
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
