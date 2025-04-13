// Package testutils provides utilities for testing the config package.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// MaxEnvParts is the maximum number of parts in an environment variable
	MaxEnvParts = 2
)

// SetupConfigTestEnv sets up a test environment for config tests
func SetupConfigTestEnv(t *testing.T) func() {
	// Store original environment
	origEnv := make(map[string]string)
	for _, env := range os.Environ() {
		origEnv[env] = os.Getenv(env)
	}

	// Clear environment
	os.Clearenv()

	// Set test environment variables
	t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	t.Setenv("GOCRAWL_CONFIG_FILE", filepath.Join("testdata", "config.yml"))
	t.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", filepath.Join("testdata", "sources.yml"))

	// Return cleanup function
	return func() {
		// Restore original environment
		os.Clearenv()
		for k, v := range origEnv {
			os.Setenv(k, v)
		}
	}
}

// SetupConfigTestEnvWithValues sets up a test environment with custom values
func SetupConfigTestEnvWithValues(t *testing.T, values map[string]string) func() {
	// Store original environment
	origEnv := make(map[string]string)
	for _, env := range os.Environ() {
		origEnv[env] = os.Getenv(env)
	}

	// Clear environment
	os.Clearenv()

	// Set default test environment variables
	t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	t.Setenv("GOCRAWL_CONFIG_FILE", filepath.Join("testdata", "config.yml"))
	t.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", filepath.Join("testdata", "sources.yml"))

	// Set custom values
	for k, v := range values {
		t.Setenv(k, v)
	}

	// Return cleanup function
	return func() {
		// Restore original environment
		os.Clearenv()
		for k, v := range origEnv {
			os.Setenv(k, v)
		}
	}
}

// TestSetupConfigTestEnv verifies the configuration test environment setup and cleanup
func TestSetupConfigTestEnv(t *testing.T) {
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
			cleanup := SetupConfigTestEnv(t)
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
