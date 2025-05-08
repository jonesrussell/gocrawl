// Package config provides configuration management for the application.
package config

import (
	"errors"
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
		return errors.New("environment variable name cannot be empty")
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return errors.New("environment variable name cannot contain whitespace")
	}
	if strings.Contains(name, "=") {
		return errors.New("environment variable name cannot contain '='")
	}
	return nil
}

// storeEnvVars stores the current environment variables
func storeEnvVars() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", MaxEnvParts)
		if len(parts) == MaxEnvParts {
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

// setupTestEnvironment sets up the test environment with initial variables
func setupTestEnvironment(t *testing.T, initialEnv map[string]string) error {
	for k, v := range initialEnv {
		if err := validateEnvVarName(k); err != nil {
			return err
		}
		t.Setenv(k, v)
	}
	return nil
}

// verifyEnvironmentCleared checks if the environment variables are cleared
func verifyEnvironmentCleared(t *testing.T, vars map[string]string) {
	for k := range vars {
		_, exists := os.LookupEnv(k)
		require.False(t, exists, "environment should be cleared, but %s exists", k)
	}
}

// verifyEnvironmentSet checks if the environment variables are set correctly
func verifyEnvironmentSet(t *testing.T, expectedEnv map[string]string) {
	for k, v := range expectedEnv {
		actual := os.Getenv(k)
		require.Equal(t, v, actual, "environment variable %s should be set correctly", k)
	}
}

// verifyTestFiles checks if the test files are set up correctly
func verifyTestFiles(t *testing.T) {
	sourcesFile := os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
	require.NotEmpty(t, sourcesFile, "GOCRAWL_CRAWLER_SOURCE_FILE should be set")
	require.FileExists(t, sourcesFile, "sources file should exist")
}

// verifyCleanup verifies that cleanup restores the original environment
func verifyCleanup(t *testing.T, initialEnv map[string]string, cleanup func()) {
	// Set a new test variable
	t.Setenv("NEW_TEST_VAR", "new_value")

	// Run cleanup
	cleanup()

	// Verify original environment is restored
	for k, v := range initialEnv {
		actual := os.Getenv(k)
		require.Equal(t, v, actual, "original environment variable %s should be restored", k)
	}

	// Verify new test variable is cleared
	_, exists := os.LookupEnv("NEW_TEST_VAR")
	require.False(t, exists, "cleanup should have cleared new environment variables")

	// Verify test files are cleared
	sourcesFile := os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
	require.Empty(t, sourcesFile, "GOCRAWL_CRAWLER_SOURCE_FILE should be cleared")
}
