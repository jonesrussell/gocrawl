package testutils

import (
	"os"
	"testing"
)

// SetupTestEnv sets up the test environment with common environment variables
func SetupTestEnv(t *testing.T) func() {
	// Store original environment variables
	originalEnv := make(map[string]string)
	for _, key := range os.Environ() {
		originalEnv[key] = os.Getenv(key)
	}

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
	os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")

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
	// Store original environment variables
	originalEnv := make(map[string]string)
	for _, key := range os.Environ() {
		originalEnv[key] = os.Getenv(key)
	}

	// Set custom environment variables
	for key, value := range values {
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		// Restore original environment variables
		for key := range originalEnv {
			os.Setenv(key, originalEnv[key])
		}
	}
}
