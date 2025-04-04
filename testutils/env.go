package testutils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTestEnv sets up the test environment with common environment variables
func SetupTestEnv(t *testing.T) func() {
	// Store original environment variables
	originalEnv := make(map[string]string)
	for _, key := range os.Environ() {
		originalEnv[key] = os.Getenv(key)
	}

	// Get the absolute path to the testdata directory
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get caller information")

	// Walk up to find the config directory
	dir := filepath.Dir(filename)
	for filepath.Base(dir) != "gocrawl" && dir != "/" {
		dir = filepath.Dir(dir)
	}
	require.NotEqual(t, "/", dir, "failed to find gocrawl directory")

	sourcesPath := filepath.Join(dir, "internal", "config", "testdata", "sources.yml")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Set common test environment variables
	os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	os.Setenv("GOCRAWL_LOG_LEVEL", "debug")
	os.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
	os.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
	os.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
	os.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "id:test_api_key")
	os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
	os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
	os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)

	// Return cleanup function
	return func() {
		// Restore original environment variables
		for key := range originalEnv {
			os.Setenv(key, originalEnv[key])
		}
	}
}

// SetupTestEnvWithValues sets up the test environment with custom values
func SetupTestEnvWithValues(t *testing.T, values map[string]string) func() {
	// First set up the common environment
	cleanup := SetupTestEnv(t)

	// Then override with custom values
	for key, value := range values {
		os.Setenv(key, value)
	}

	return cleanup
}
