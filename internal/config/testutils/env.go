package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTestEnv sets up the test environment with configuration files
func SetupTestEnv(t *testing.T) func() {
	t.Helper()
	require := require.New(t)

	// Find the config directory
	dir, err := os.Getwd()
	require.NoError(err)

	// Walk up until we find the config directory
	for filepath.Base(dir) != "config" && dir != "/" {
		dir = filepath.Dir(dir)
	}
	require.NotEqual("/", dir, "Could not find config directory")

	// Set paths to configuration files
	configPath := filepath.Join(dir, "testutils", "testdata", "configs", "base.yml")
	sourcesPath := filepath.Join(dir, "testutils", "testdata", "configs", "sources.yml")

	// Store original environment variables
	origEnv := map[string]string{
		"GOCRAWL_CONFIG_FILE":                      os.Getenv("GOCRAWL_CONFIG_FILE"),
		"GOCRAWL_SOURCES_FILE":                     os.Getenv("GOCRAWL_SOURCES_FILE"),
		"GOCRAWL_APP_ENVIRONMENT":                  os.Getenv("GOCRAWL_APP_ENVIRONMENT"),
		"GOCRAWL_APP_NAME":                         os.Getenv("GOCRAWL_APP_NAME"),
		"GOCRAWL_APP_VERSION":                      os.Getenv("GOCRAWL_APP_VERSION"),
		"GOCRAWL_LOG_LEVEL":                        os.Getenv("GOCRAWL_LOG_LEVEL"),
		"GOCRAWL_CRAWLER_BASE_URL":                 os.Getenv("GOCRAWL_CRAWLER_BASE_URL"),
		"GOCRAWL_CRAWLER_MAX_DEPTH":                os.Getenv("GOCRAWL_CRAWLER_MAX_DEPTH"),
		"GOCRAWL_CRAWLER_PARALLELISM":              os.Getenv("GOCRAWL_CRAWLER_PARALLELISM"),
		"GOCRAWL_CRAWLER_RATE_LIMIT":               os.Getenv("GOCRAWL_CRAWLER_RATE_LIMIT"),
		"GOCRAWL_CRAWLER_SOURCE_FILE":              os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE"),
		"GOCRAWL_SERVER_SECURITY_ENABLED":          os.Getenv("GOCRAWL_SERVER_SECURITY_ENABLED"),
		"GOCRAWL_SERVER_SECURITY_API_KEY":          os.Getenv("GOCRAWL_SERVER_SECURITY_API_KEY"),
		"GOCRAWL_ELASTICSEARCH_ADDRESSES":          os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"),
		"GOCRAWL_ELASTICSEARCH_INDEX_NAME":         os.Getenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME"),
		"GOCRAWL_ELASTICSEARCH_API_KEY":            os.Getenv("GOCRAWL_ELASTICSEARCH_API_KEY"),
		"GOCRAWL_ELASTICSEARCH_TLS_ENABLED":        os.Getenv("GOCRAWL_ELASTICSEARCH_TLS_ENABLED"),
		"GOCRAWL_ELASTICSEARCH_RETRY_ENABLED":      os.Getenv("GOCRAWL_ELASTICSEARCH_RETRY_ENABLED"),
		"GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT": os.Getenv("GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT"),
		"GOCRAWL_ELASTICSEARCH_RETRY_MAX_WAIT":     os.Getenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_WAIT"),
		"GOCRAWL_ELASTICSEARCH_RETRY_MAX_RETRIES":  os.Getenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_RETRIES"),
		"GOCRAWL_ELASTICSEARCH_BULK_SIZE":          os.Getenv("GOCRAWL_ELASTICSEARCH_BULK_SIZE"),
		"GOCRAWL_ELASTICSEARCH_FLUSH_INTERVAL":     os.Getenv("GOCRAWL_ELASTICSEARCH_FLUSH_INTERVAL"),
	}

	// Set test environment variables
	envVars := map[string]string{
		"GOCRAWL_CONFIG_FILE":                      configPath,
		"GOCRAWL_SOURCES_FILE":                     sourcesPath,
		"GOCRAWL_APP_ENVIRONMENT":                  "test",
		"GOCRAWL_APP_NAME":                         "gocrawl-test",
		"GOCRAWL_APP_VERSION":                      "0.0.1",
		"GOCRAWL_LOG_LEVEL":                        "debug",
		"GOCRAWL_CRAWLER_BASE_URL":                 "http://test.example.com",
		"GOCRAWL_CRAWLER_MAX_DEPTH":                "2",
		"GOCRAWL_CRAWLER_PARALLELISM":              "2",
		"GOCRAWL_CRAWLER_RATE_LIMIT":               "2s",
		"GOCRAWL_CRAWLER_SOURCE_FILE":              sourcesPath,
		"GOCRAWL_SERVER_SECURITY_ENABLED":          "true",
		"GOCRAWL_SERVER_SECURITY_API_KEY":          "id:test_api_key",
		"GOCRAWL_ELASTICSEARCH_ADDRESSES":          "http://localhost:9200",
		"GOCRAWL_ELASTICSEARCH_INDEX_NAME":         "test-index",
		"GOCRAWL_ELASTICSEARCH_API_KEY":            "id:test_api_key",
		"GOCRAWL_ELASTICSEARCH_TLS_ENABLED":        "false",
		"GOCRAWL_ELASTICSEARCH_RETRY_ENABLED":      "true",
		"GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT": "1s",
		"GOCRAWL_ELASTICSEARCH_RETRY_MAX_WAIT":     "5s",
		"GOCRAWL_ELASTICSEARCH_RETRY_MAX_RETRIES":  "3",
		"GOCRAWL_ELASTICSEARCH_BULK_SIZE":          "1000",
		"GOCRAWL_ELASTICSEARCH_FLUSH_INTERVAL":     "30s",
	}

	// Set new environment variables
	for k, v := range envVars {
		err = os.Setenv(k, v)
		require.NoError(err)
	}

	// Return cleanup function
	return func() {
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
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

	// Clear all environment variables
	for key := range originalEnv {
		os.Unsetenv(key)
	}

	// Set custom environment variables
	for key, value := range values {
		os.Setenv(key, value)
	}

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
