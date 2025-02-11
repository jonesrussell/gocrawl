package config_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	// Setup test environment variables
	testEnv := map[string]string{
		"APP_ENV":            "test",
		"LOG_LEVEL":          "debug",
		"APP_DEBUG":          "true",
		"ELASTIC_URL":        "http://test-elastic:9200",
		"ELASTIC_PASSWORD":   "test-password",
		"ELASTIC_API_KEY":    "test-key",
		"CRAWLER_BASE_URL":   "http://test.com",
		"CRAWLER_MAX_DEPTH":  "3",
		"CRAWLER_RATE_LIMIT": "2s",
		"INDEX_NAME":         "test-index",
	}

	// Set environment variables
	for k, v := range testEnv {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Test successful config creation
	t.Run("successful config creation", func(t *testing.T) {
		cfg, err := config.NewConfig(http.DefaultTransport)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Verify app config
		assert.Equal(t, "test", cfg.App.Environment)
		assert.Equal(t, "debug", cfg.App.LogLevel)
		assert.True(t, cfg.App.Debug)

		// Verify crawler config
		assert.Equal(t, "http://test.com", cfg.Crawler.BaseURL)
		assert.Equal(t, 3, cfg.Crawler.MaxDepth)
		assert.Equal(t, 2*time.Second, cfg.Crawler.RateLimit)
		assert.Equal(t, "test-index", cfg.Crawler.IndexName)
		assert.NotNil(t, cfg.Crawler.Transport)

		// Verify elasticsearch config
		assert.Equal(t, "http://test-elastic:9200", cfg.Elasticsearch.URL)
		assert.Equal(t, "test-password", cfg.Elasticsearch.Password)
		assert.Equal(t, "test-key", cfg.Elasticsearch.APIKey)
	})

	// Test missing required fields
	t.Run("missing elastic url", func(t *testing.T) {
		os.Unsetenv("ELASTIC_URL")
		cfg, err := config.NewConfig(http.DefaultTransport)
		assert.ErrorIs(t, err, config.ErrMissingElasticURL)
		assert.Nil(t, cfg)
	})

	// Test default values
	t.Run("default values", func(t *testing.T) {
		// Clear all environment variables
		for k := range testEnv {
			os.Unsetenv(k)
		}
		// Set only required ELASTIC_URL
		os.Setenv("ELASTIC_URL", "http://localhost:9200")

		cfg, err := config.NewConfig(http.DefaultTransport)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Verify default values
		assert.Equal(t, "development", cfg.App.Environment)
		assert.Equal(t, "info", cfg.App.LogLevel)
		assert.False(t, cfg.App.Debug)
		assert.Equal(t, time.Second, cfg.Crawler.RateLimit)
		assert.Equal(t, 0, cfg.Crawler.MaxDepth)
		assert.Equal(t, "", cfg.Crawler.IndexName)
	})
}

func TestLoadConfig_Success(t *testing.T) {
	// Set environment variables for testing
	defer os.Clearenv() // Clear environment variables after the test
	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ELASTIC_URL", "http://localhost:9200")
	t.Setenv("ELASTIC_PASSWORD", "password")
	t.Setenv("ELASTIC_API_KEY", "api_key")
	t.Setenv("INDEX_NAME", "test_index")
	t.Setenv("CRAWLER_BASE_URL", "https://example.com")
	t.Setenv("CRAWLER_MAX_DEPTH", "3")

	// Load the configuration
	cfg, err := config.NewConfig(http.DefaultTransport)
	require.NoError(t, err)

	// Assert the loaded values
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, "http://localhost:9200", cfg.Elasticsearch.URL)
	assert.Equal(t, "password", cfg.Elasticsearch.Password)
	assert.Equal(t, "api_key", cfg.Elasticsearch.APIKey)
	assert.Equal(t, "test_index", cfg.Crawler.IndexName)
	assert.Equal(t, "https://example.com", cfg.Crawler.BaseURL)
	assert.Equal(t, 3, cfg.Crawler.MaxDepth)
}

func TestLoadConfig_MissingElasticURL(t *testing.T) {
	defer os.Clearenv() // Clear environment variables after the test
	// Set environment variables for testing
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_DEBUG", "true")
	t.Setenv("ELASTIC_PASSWORD", "password")
	t.Setenv("ELASTIC_API_KEY", "api_key")
	t.Setenv("INDEX_NAME", "test_index")
	t.Setenv("LOG_LEVEL", "debug")

	// Load the configuration
	cfg, err := config.NewConfig(http.DefaultTransport)

	// Assert that an error is returned and ElasticURL is required
	require.Error(t, err)
	assert.Nil(t, cfg)
	require.EqualError(t, err, "ELASTIC_URL is required")
}

func TestLoadConfig_EnvFileNotFound(t *testing.T) {
	defer os.Clearenv() // Clear environment variables after the test
	// Ensure that the .env file is not loaded
	t.Setenv("APP_ENV", "development")
	// Do not set ELASTIC_URL to simulate the missing environment variable

	// Load the configuration
	cfg, err := config.NewConfig(http.DefaultTransport)

	// Assert that an error is returned for the missing ELASTIC_URL
	require.Error(t, err)
	require.EqualError(t, err, "ELASTIC_URL is required")
	assert.Nil(t, cfg) // Ensure that the config is nil
}
