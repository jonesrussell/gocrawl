package config_test

import (
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Set environment variables for testing
	defer os.Clearenv() // Clear environment variables after the test
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_DEBUG", "true")
	t.Setenv("ELASTIC_URL", "http://localhost:9200")
	t.Setenv("ELASTIC_PASSWORD", "password")
	t.Setenv("ELASTIC_API_KEY", "api_key")
	t.Setenv("INDEX_NAME", "test_index")
	t.Setenv("LOG_LEVEL", "debug")

	// Load the configuration
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	// Assert the loaded values
	assert.Equal(t, "TestApp", cfg.AppName)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.True(t, cfg.AppDebug)
	assert.Equal(t, "http://localhost:9200", cfg.ElasticURL)
	assert.Equal(t, "password", cfg.ElasticPassword)
	assert.Equal(t, "api_key", cfg.ElasticAPIKey)
	assert.Equal(t, "test_index", cfg.IndexName)
	assert.Equal(t, "debug", cfg.LogLevel)
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
	cfg, err := config.LoadConfig()

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
	cfg, err := config.LoadConfig()

	// Assert that an error is returned for the missing ELASTIC_URL
	require.Error(t, err)
	require.EqualError(t, err, "ELASTIC_URL is required")
	assert.Nil(t, cfg) // Ensure that the config is nil
}
