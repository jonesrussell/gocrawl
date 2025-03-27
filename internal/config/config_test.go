package config_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestNew(t *testing.T) {
	// Save current environment and use t.Setenv for automatic cleanup
	t.Setenv("APP_ENV", "")

	// Reset Viper after each test
	defer viper.Reset()

	// Create test config file
	testConfig := `
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
  source_file: ./testdata/sources.yml
log:
  level: debug
  debug: true
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  tls:
    skip_verify: true
`
	writeErr := os.WriteFile("testdata/config.yml", []byte(testConfig), 0644)
	require.NoError(t, writeErr)
	defer func() {
		if removeErr := os.Remove("testdata/config.yml"); removeErr != nil {
			t.Errorf("Error removing config file: %v", removeErr)
		}
	}()

	// Create test sources file
	testSources := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        content: article
        author: .author
        date: .date
`
	writeErr = os.WriteFile("testdata/sources.yml", []byte(testSources), 0644)
	require.NoError(t, writeErr)
	defer func() {
		if removeErr := os.Remove("testdata/sources.yml"); removeErr != nil {
			t.Errorf("Error removing sources file: %v", removeErr)
		}
	}()

	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Reset viper for clean state
				viper.Reset()

				// Set test environment
				t.Setenv("APP_ENV", "test")
				t.Setenv("CONFIG_FILE", "./testdata/config.yml")
				t.Setenv("LOG_LEVEL", "debug")

				// Configure viper
				viper.SetConfigFile("./testdata/config.yml")
				viper.AutomaticEnv()
				viper.SetEnvPrefix("")
				viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

				// Bind environment variables
				require.NoError(t, viper.BindEnv("app.environment", "APP_ENV"))
				require.NoError(t, viper.BindEnv("log.level", "LOG_LEVEL"))
				require.NoError(t, viper.BindEnv("elasticsearch.api_key", "ELASTIC_API_KEY"))

				// Read config file
				readErr := viper.ReadInConfig()
				require.NoError(t, readErr)
			},
			validate: func(t *testing.T, cfg config.Interface, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)

				appCfg := cfg.GetAppConfig()
				require.Equal(t, "test", appCfg.Environment)

				logCfg := cfg.GetLogConfig()
				require.Equal(t, "debug", logCfg.Level)
				require.True(t, logCfg.Debug)

				crawlerCfg := cfg.GetCrawlerConfig()
				require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
				require.Equal(t, 2, crawlerCfg.MaxDepth)
				require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
				require.Equal(t, 2, crawlerCfg.Parallelism)

				elasticCfg := cfg.GetElasticsearchConfig()
				require.Equal(t, []string{"https://localhost:9200"}, elasticCfg.Addresses)
				require.Equal(t, "", elasticCfg.Username)
				require.Equal(t, "", elasticCfg.Password)
				require.Equal(t, "test_api_key", elasticCfg.APIKey)
				require.True(t, elasticCfg.TLS.SkipVerify)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
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

func TestLoadConfig(t *testing.T) {
	// Ensure testdata directory exists
	err := os.MkdirAll("internal/config/testdata", 0755)
	require.NoError(t, err)

	// Create valid config file for testing
	validConfig := `
app:
  environment: development
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  max_depth: 3
  parallelism: 2
  source_file: internal/config/testdata/sources.yml
log:
  level: debug
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: test_key
`
	err = os.WriteFile("internal/config/testdata/config.yml", []byte(validConfig), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/config.yml"); removeErr != nil {
			t.Errorf("Error removing test file: %v", removeErr)
		}
	}()

	// Create test sources file
	testSources := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        content: article
        author: .author
        date: .date
`
	err = os.WriteFile("internal/config/testdata/sources.yml", []byte(testSources), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/sources.yml"); removeErr != nil {
			t.Errorf("Error removing sources file: %v", removeErr)
		}
	}()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid config file",
			path:    "internal/config/testdata/config.yml",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "internal/config/testdata/nonexistent.yml",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			path:    "internal/config/testdata/invalid.yml",
			wantErr: true,
		},
	}

	// Create invalid YAML file for testing
	invalidYAML := []byte("invalid: yaml: content")
	err = os.WriteFile("internal/config/testdata/invalid.yml", invalidYAML, 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/invalid.yml"); removeErr != nil {
			t.Errorf("Error removing test file: %v", removeErr)
		}
	}()

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

func TestConfigurationPriority(t *testing.T) {
	// Save current environment and restore after test
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range originalEnv {
			k, v, _ := strings.Cut(e, "=")
			t.Setenv(k, v)
		}
		viper.Reset()
	}()

	// Clear environment and viper config
	os.Clearenv()
	viper.Reset()

	// Create test .env file
	envContent := `
APP_ENV=development
LOG_LEVEL=info
ELASTIC_API_KEY=env_api_key
`
	envErr := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, envErr)
	defer func() {
		if removeErr := os.Remove(".env"); removeErr != nil {
			t.Errorf("Error removing .env file: %v", removeErr)
		}
	}()

	// Create test config files
	configWithAPIKey := `
app:
  environment: production
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: test_api_key
log:
  level: debug
crawler:
  max_depth: 3
  parallelism: 2
  rate_limit: 1s
  source_file: ./testdata/sources.yml
`
	err := os.WriteFile("testdata/config.yml", []byte(configWithAPIKey), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("testdata/config.yml"); removeErr != nil {
			t.Errorf("Error removing config file: %v", removeErr)
		}
	}()

	// Create test sources file
	testSources := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        content: article
        author: .author
        date: .date
`
	err = os.WriteFile("testdata/sources.yml", []byte(testSources), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("testdata/sources.yml"); removeErr != nil {
			t.Errorf("Error removing sources file: %v", removeErr)
		}
	}()

	// Set up Viper
	viper.SetConfigFile("testdata/config.yml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	require.NoError(t, viper.BindEnv("elasticsearch.api_key", "ELASTIC_API_KEY"))
	require.NoError(t, viper.BindEnv("app.environment", "APP_ENV"))
	require.NoError(t, viper.BindEnv("log.level", "LOG_LEVEL"))

	// Read config file
	require.NoError(t, viper.ReadInConfig())

	// Create config
	cfg, err := config.New()
	require.NoError(t, err)

	// Verify environment variables take precedence over config file values
	require.Equal(t, "env_api_key", cfg.GetElasticsearchConfig().APIKey,
		"Environment variable ELASTIC_API_KEY should override config file value")
	require.Equal(t, "development", cfg.GetAppConfig().Environment,
		"Environment variable APP_ENV should override config file value")
	require.Equal(t, "info", cfg.GetLogConfig().Level,
		"Environment variable LOG_LEVEL should override config file value")
}

func TestRequiredConfigurationValidation(t *testing.T) {
	// Create test .env file
	envContent := `
APP_ENV=development
LOG_LEVEL=debug
ELASTIC_API_KEY=test_key
`
	envErr := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, envErr)
	defer func() {
		if removeErr := os.Remove(".env"); removeErr != nil {
			t.Errorf("Error removing .env file: %v", removeErr)
		}
	}()

	// Create test config files
	validConfig := `
app:
  environment: development
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  max_depth: 3
  parallelism: 2
  source_file: ./testdata/sources.yml
log:
  level: debug
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: test_key
`
	noAPIKeyConfig := `
app:
  environment: production
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  max_depth: 3
  parallelism: 2
  source_file: ./testdata/sources.yml
log:
  level: debug
elasticsearch:
  addresses:
    - http://localhost:9200
`
	writeErr1 := os.WriteFile("testdata/valid_config.yml", []byte(validConfig), 0644)
	require.NoError(t, writeErr1)
	defer func() {
		if removeErr := os.Remove("testdata/valid_config.yml"); removeErr != nil {
			t.Errorf("Error removing valid config file: %v", removeErr)
		}
	}()

	writeErr2 := os.WriteFile("testdata/no_api_key_config.yml", []byte(noAPIKeyConfig), 0644)
	require.NoError(t, writeErr2)
	defer func() {
		if removeErr := os.Remove("testdata/no_api_key_config.yml"); removeErr != nil {
			t.Errorf("Error removing no API key config file: %v", removeErr)
		}
	}()

	// Create test sources file
	testSources := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        content: article
        author: .author
        date: .date
`
	writeErr3 := os.WriteFile("testdata/sources.yml", []byte(testSources), 0644)
	require.NoError(t, writeErr3)
	defer func() {
		if removeErr := os.Remove("testdata/sources.yml"); removeErr != nil {
			t.Errorf("Error removing sources file: %v", removeErr)
		}
	}()

	// Create test environment
	tests := []struct {
		name        string
		envVars     map[string]string
		configFile  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration with all required fields",
			envVars: map[string]string{
				"ELASTIC_API_KEY": "test_key",
				"APP_ENV":         "development",
				"LOG_LEVEL":       "debug",
				"CONFIG_FILE":     "testdata/valid_config.yml",
			},
			configFile:  "testdata/valid_config.yml",
			expectError: false,
		},
		{
			name: "missing API key in production",
			envVars: map[string]string{
				"APP_ENV":     "production",
				"LOG_LEVEL":   "debug",
				"CONFIG_FILE": "testdata/no_api_key_config.yml",
			},
			configFile:  "testdata/no_api_key_config.yml",
			expectError: true,
			errorMsg:    "API key is required in production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Configure viper
			viper.SetConfigFile(tt.configFile)
			viper.AutomaticEnv()
			viper.SetEnvPrefix("") // No prefix for env vars
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			// Set up environment variable bindings
			require.NoError(t, viper.BindEnv("app.environment", "APP_ENV"))
			require.NoError(t, viper.BindEnv("elasticsearch.api_key", "ELASTIC_API_KEY"))
			require.NoError(t, viper.BindEnv("log.level", "LOG_LEVEL"))

			// Read config file
			readErr := viper.ReadInConfig()
			require.NoError(t, readErr)

			// Initialize config
			cfg, initErr := config.New()

			if tt.expectError {
				require.Error(t, initErr)
				assert.Contains(t, initErr.Error(), tt.errorMsg)
			} else {
				require.NoError(t, initErr)
				require.NotNil(t, cfg)

				// Verify configuration
				appCfg := cfg.GetAppConfig()
				require.Equal(t, tt.envVars["APP_ENV"], appCfg.Environment)

				esCfg := cfg.GetElasticsearchConfig()
				require.Equal(t, tt.envVars["ELASTIC_API_KEY"], esCfg.APIKey)

				logCfg := cfg.GetLogConfig()
				require.Equal(t, tt.envVars["LOG_LEVEL"], logCfg.Level)
			}
		})
	}
}
