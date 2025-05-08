// Package config provides configuration management for the application.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
