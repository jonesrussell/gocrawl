// Package testutils provides utilities for testing the config package.
package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// MaxEnvParts is the maximum number of parts in an environment variable
	MaxEnvParts = 2
)

// defaultEnvVars defines the default environment variables for testing
var defaultEnvVars = map[string]string{
	"GOCRAWL_APP_ENVIRONMENT":     "test",
	"GOCRAWL_APP_NAME":            "gocrawl-test",
	"GOCRAWL_APP_VERSION":         "0.0.1",
	"GOCRAWL_CONFIG_FILE":         filepath.Join("testdata", "config", "base.yml"),
	"GOCRAWL_CRAWLER_SOURCE_FILE": filepath.Join("testdata", "sources", "basic.yml"),
}

// validateEnvVarName checks if an environment variable name is valid
func validateEnvVarName(name string) error {
	if name == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("environment variable name cannot contain whitespace")
	}
	if strings.Contains(name, "=") {
		return fmt.Errorf("environment variable name cannot contain '='")
	}
	return nil
}

// storeEnvVars stores the current environment variables
func storeEnvVars() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// restoreEnvVars restores environment variables from a map
func restoreEnvVars(t *testing.T, env map[string]string) {
	os.Clearenv()
	for k, v := range env {
		if err := validateEnvVarName(k); err != nil {
			t.Logf("warning: skipping invalid environment variable %q: %v", k, err)
			continue
		}
		t.Setenv(k, v)
	}
}

// SetupConfigTestEnv sets up a test environment for config tests
func SetupConfigTestEnv(t *testing.T) func() {
	t.Helper()

	// Store original environment
	origEnv := storeEnvVars()

	// Clear environment
	os.Clearenv()

	// Set test environment variables
	for k, v := range defaultEnvVars {
		if err := validateEnvVarName(k); err != nil {
			t.Fatalf("invalid environment variable name %q: %v", k, err)
		}
		t.Setenv(k, v)
	}

	// Return cleanup function
	return func() {
		restoreEnvVars(t, origEnv)
	}
}

// SetupConfigTestEnvWithValues sets up a test environment with custom values
func SetupConfigTestEnvWithValues(t *testing.T, values map[string]string) func() {
	t.Helper()

	// Store original environment
	origEnv := storeEnvVars()

	// Clear environment
	os.Clearenv()

	// Set default test environment variables
	for k, v := range defaultEnvVars {
		if err := validateEnvVarName(k); err != nil {
			t.Fatalf("invalid environment variable name %q: %v", k, err)
		}
		t.Setenv(k, v)
	}

	// Set custom values
	for k, v := range values {
		if err := validateEnvVarName(k); err != nil {
			t.Fatalf("invalid environment variable name %q: %v", k, err)
		}
		t.Setenv(k, v)
	}

	// Return cleanup function
	return func() {
		restoreEnvVars(t, origEnv)
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
		wantErr       bool
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
		{
			name: "invalid environment variable name",
			initialEnv: map[string]string{
				"INVALID=VAR": "test_value",
			},
			wantErr: true,
		},
		{
			name: "empty environment variable name",
			initialEnv: map[string]string{
				"": "test_value",
			},
			wantErr: true,
		},
		{
			name: "whitespace in environment variable name",
			initialEnv: map[string]string{
				"TEST VAR": "test_value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup environment variables
			for k, v := range tt.initialEnv {
				if err := validateEnvVarName(k); err != nil {
					if !tt.wantErr {
						t.Fatalf("unexpected error: %v", err)
					}
					return
				}
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
