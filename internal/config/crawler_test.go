package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

// createTestConfig creates a test configuration file with the given values
func createTestConfig(t *testing.T, values map[string]interface{}) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	sourcesPath := filepath.Join(tmpDir, "sources.yml")

	// Create test config file
	configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: ` + sourcesPath + `
log:
  level: debug
  debug: true
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  index_name: test-index
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
server:
  security:
    enabled: true
    api_key: id:test_api_key
`

	// Override values from the map
	for key, value := range values {
		viper.Set(key, value)
	}

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create test sources file
	sourcesContent := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        body: article
        author: .author
        published_time: .date
`
	err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
	require.NoError(t, err)

	return configPath
}

func TestCrawlerConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string)
		validate    func(t *testing.T, cfg *config.CrawlerConfig)
		expectError bool
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) (string, string) {
				// Create temporary config file
				configPath := filepath.Join(t.TempDir(), "config.yml")
				require.NoError(t, os.WriteFile(configPath, []byte(`
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
log:
  level: info
  debug: false
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
server:
  address: :8080
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: sources.yml
`), 0644))

				// Create temporary sources file
				sourcesPath := filepath.Join(t.TempDir(), "sources.yml")
				require.NoError(t, os.WriteFile(sourcesPath, []byte(`
sources:
  - name: test
    url: http://test.example.com
    selectors:
      article: article
      title: h1
      content: .content
`), 0644))

				// Set environment variables
				testutils.SetupTestEnvWithValues(t, map[string]string{
					"GOCRAWL_APP_ENVIRONMENT":     "test",
					"GOCRAWL_CRAWLER_SOURCE_FILE": sourcesPath,
					"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
					"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
					"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
					"GOCRAWL_CRAWLER_PARALLELISM": "2",
				})

				return configPath, sourcesPath
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.NotNil(t, cfg)
				require.Equal(t, "http://test.example.com", cfg.BaseURL)
				require.Equal(t, 2, cfg.MaxDepth)
				require.Equal(t, "2s", cfg.RateLimit.String())
				require.Equal(t, 2, cfg.Parallelism)
			},
			expectError: false,
		},
		{
			name: "invalid max depth",
			setup: func(t *testing.T) (string, string) {
				// Create temporary config file
				configPath := filepath.Join(t.TempDir(), "config.yml")
				require.NoError(t, os.WriteFile(configPath, []byte(`
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
log:
  level: info
  debug: false
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
server:
  address: :8080
crawler:
  base_url: http://test.example.com
  max_depth: 0
  rate_limit: 2s
  parallelism: 2
  source_file: sources.yml
`), 0644))

				// Create temporary sources file
				sourcesPath := filepath.Join(t.TempDir(), "sources.yml")
				require.NoError(t, os.WriteFile(sourcesPath, []byte(`
sources:
  - name: test
    url: http://test.example.com
    selectors:
      article: article
      title: h1
      content: .content
`), 0644))

				// Set environment variables
				testutils.SetupTestEnvWithValues(t, map[string]string{
					"GOCRAWL_APP_ENVIRONMENT":     "test",
					"GOCRAWL_CRAWLER_SOURCE_FILE": sourcesPath,
					"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
					"GOCRAWL_CRAWLER_MAX_DEPTH":   "0",
					"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
					"GOCRAWL_CRAWLER_PARALLELISM": "2",
				})

				return configPath, sourcesPath
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Nil(t, cfg)
			},
			expectError: true,
		},
		{
			name: "invalid parallelism",
			setup: func(t *testing.T) (string, string) {
				// Create temporary config file
				configPath := filepath.Join(t.TempDir(), "config.yml")
				require.NoError(t, os.WriteFile(configPath, []byte(`
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
log:
  level: info
  debug: false
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
server:
  address: :8080
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 0
  source_file: sources.yml
`), 0644))

				// Create temporary sources file
				sourcesPath := filepath.Join(t.TempDir(), "sources.yml")
				require.NoError(t, os.WriteFile(sourcesPath, []byte(`
sources:
  - name: test
    url: http://test.example.com
    selectors:
      article: article
      title: h1
      content: .content
`), 0644))

				// Set environment variables
				testutils.SetupTestEnvWithValues(t, map[string]string{
					"GOCRAWL_APP_ENVIRONMENT":     "test",
					"GOCRAWL_CRAWLER_SOURCE_FILE": sourcesPath,
					"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
					"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
					"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
					"GOCRAWL_CRAWLER_PARALLELISM": "0",
				})

				return configPath, sourcesPath
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Nil(t, cfg)
			},
			expectError: true,
		},
		{
			name: "invalid rate limit",
			setup: func(t *testing.T) (string, string) {
				// Create temporary config file
				configPath := filepath.Join(t.TempDir(), "config.yml")
				require.NoError(t, os.WriteFile(configPath, []byte(`
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
log:
  level: info
  debug: false
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
server:
  address: :8080
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: invalid
  parallelism: 2
  source_file: sources.yml
`), 0644))

				// Create temporary sources file
				sourcesPath := filepath.Join(t.TempDir(), "sources.yml")
				require.NoError(t, os.WriteFile(sourcesPath, []byte(`
sources:
  - name: test
    url: http://test.example.com
    selectors:
      article: article
      title: h1
      content: .content
`), 0644))

				// Set environment variables
				testutils.SetupTestEnvWithValues(t, map[string]string{
					"GOCRAWL_APP_ENVIRONMENT":     "test",
					"GOCRAWL_CRAWLER_SOURCE_FILE": sourcesPath,
					"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
					"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
					"GOCRAWL_CRAWLER_RATE_LIMIT":  "invalid",
					"GOCRAWL_CRAWLER_PARALLELISM": "2",
				})

				return configPath, sourcesPath
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Nil(t, cfg)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			configPath, sourcesPath := tt.setup(t)
			defer os.Remove(configPath)
			defer os.Remove(sourcesPath)

			// Set config file path in environment
			os.Setenv("GOCRAWL_CONFIG", configPath)
			defer os.Unsetenv("GOCRAWL_CONFIG")

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.expectError {
				require.Error(t, err)
				if tt.validate != nil {
					tt.validate(t, nil)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.validate != nil {
					tt.validate(t, cfg.GetCrawlerConfig())
				}
			}
		})
	}
}

func TestCrawlerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T)
		wantErr bool
	}{
		{
			name: "invalid max depth",
			setup: func(t *testing.T) {
				// Create test config file with invalid max depth
				configContent := `
crawler:
  base_url: http://test.example.com
  max_depth: 0
  rate_limit: 2s
  parallelism: 2
  source_file: internal/config/testdata/sources.yml
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("GOCRAWL_CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "invalid parallelism",
			setup: func(t *testing.T) {
				// Create test config file with invalid parallelism
				configContent := `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 0
  source_file: internal/config/testdata/sources.yml
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("GOCRAWL_CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			setup: func(t *testing.T) {
				// Create test config file with invalid rate limit
				configContent := `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 0s
  parallelism: 2
  source_file: internal/config/testdata/sources.yml
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("GOCRAWL_CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := testutils.SetupTestEnv(t)
			defer cleanup()

			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.New(testutils.NewTestLogger(t))

			// Validate results
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
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
			require.Equal(t, tt.expected, duration)
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
