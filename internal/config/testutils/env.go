package testutils

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTestEnv sets up the test environment with common environment variables
func SetupTestEnv(t *testing.T) func() {
	// Store original environment variables
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Clear all environment variables
	for key := range originalEnv {
		os.Unsetenv(key)
	}

	// Set common test environment variables
	os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	os.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	os.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	os.Setenv("GOCRAWL_LOG_LEVEL", "debug")
	os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
	os.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
	os.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
	os.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
	os.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "id:test_api_key")
	os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
	os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
	os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "../testdata/sources.yml")

	// Return cleanup function
	return func() {
		// Clear all current environment variables
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) == 2 {
				os.Unsetenv(pair[0])
			}
		}

		// Restore original environment variables
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}
}

// SetupTestEnvWithValues sets up the test environment with custom values
func SetupTestEnvWithValues(t *testing.T, values map[string]string) func() {
	// Store original environment variables
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Set custom environment variables
	for key, value := range values {
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		// Restore original environment variables
		for key, value := range originalEnv {
			os.Setenv(key, value)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set initial environment
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
