package testutils

import (
	"os"
	"strings"
	"testing"
)

const (
	// MaxEnvParts is the maximum number of parts in an environment variable
	MaxEnvParts = 2
)

// SetupTestEnv sets up the test environment with default values
func SetupTestEnv(t *testing.T) func() {
	t.Helper()

	// Store original environment
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", MaxEnvParts)
		if len(pair) == MaxEnvParts {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Set test environment variables
	for key, value := range map[string]string{
		"GOCRAWL_APP_ENVIRONMENT":         "test",
		"GOCRAWL_LOG_LEVEL":               "debug",
		"GOCRAWL_CRAWLER_MAX_DEPTH":       "2",
		"GOCRAWL_CRAWLER_PARALLELISM":     "2",
		"GOCRAWL_SERVER_SECURITY_ENABLED": "true",
		"GOCRAWL_SERVER_SECURITY_API_KEY": "id:test_api_key",
	} {
		t.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, value := range originalEnv {
			t.Setenv(key, value)
		}
	}
}

// SetupTestEnvWithValues sets up the test environment with custom values
func SetupTestEnvWithValues(t *testing.T, values map[string]string) func() {
	t.Helper()

	// Store original environment
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", MaxEnvParts)
		if len(pair) == MaxEnvParts {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Set test environment variables
	for key, value := range values {
		t.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, value := range originalEnv {
			t.Setenv(key, value)
		}
	}
}
