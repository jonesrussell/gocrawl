package testutils

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTestEnv sets up the test environment with default values
func SetupTestEnv(t *testing.T) func() {
	t.Helper()

	// Store original environment
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Set test environment variables
	for key, value := range map[string]string{
		"GOCRAWL_APP_ENVIRONMENT":          "test",
		"GOCRAWL_LOG_LEVEL":                "debug",
		"GOCRAWL_CRAWLER_MAX_DEPTH":        "2",
		"GOCRAWL_CRAWLER_PARALLELISM":      "2",
		"GOCRAWL_SERVER_SECURITY_ENABLED":  "true",
		"GOCRAWL_SERVER_SECURITY_API_KEY":  "id:test_api_key",
		"GOCRAWL_ELASTICSEARCH_ADDRESSES":  "http://localhost:9200",
		"GOCRAWL_ELASTICSEARCH_INDEX_NAME": "test-index",
		"GOCRAWL_ELASTICSEARCH_API_KEY":    "id:test_api_key",
		"GOCRAWL_CRAWLER_SOURCE_FILE":      "testdata/sources.yml",
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
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
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

// TestSetupTestEnv verifies the test environment setup and cleanup
func TestSetupTestEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialEnv    map[string]string
		expectedEnv   map[string]string
		verifyCleanup bool
	}{
		{
			name: "basic environment setup and cleanup",
			initialEnv: map[string]string{
				"TEST_VAR": "test_value",
			},
			expectedEnv: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT": "test",
				"GOCRAWL_APP_NAME":        "gocrawl-test",
				"GOCRAWL_APP_VERSION":     "0.0.1",
			},
			verifyCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup environment variables
			for k, v := range tt.initialEnv {
				t.Setenv(k, v)
			}

			// Setup test environment
			cleanup := SetupTestEnv(t)
			defer cleanup()

			// Verify environment is cleared
			for k := range tt.initialEnv {
				_, exists := os.LookupEnv(k)
				require.False(t, exists, "environment should be cleared, but %s exists", k)
			}

			// Verify test environment is set correctly
			for k, v := range tt.expectedEnv {
				actual := os.Getenv(k)
				require.Equal(t, v, actual, "environment variable %s should be set correctly", k)
			}

			// Verify test files are set
			sourcesFile := os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
			require.NotEmpty(t, sourcesFile, "GOCRAWL_CRAWLER_SOURCE_FILE should be set")
			require.FileExists(t, sourcesFile, "sources file should exist")

			if tt.verifyCleanup {
				// Set a new test variable
				t.Setenv("NEW_TEST_VAR", "new_value")

				// Run cleanup
				cleanup()

				// Verify original environment is restored
				for k, v := range tt.initialEnv {
					actual := os.Getenv(k)
					require.Equal(t, v, actual, "original environment variable %s should be restored", k)
				}

				// Verify new test variable is cleared
				_, exists := os.LookupEnv("NEW_TEST_VAR")
				require.False(t, exists, "cleanup should have cleared new environment variables")

				// Verify test files are cleared
				sourcesFile = os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
				require.Empty(t, sourcesFile, "GOCRAWL_CRAWLER_SOURCE_FILE should be cleared")
			}
		})
	}
}
