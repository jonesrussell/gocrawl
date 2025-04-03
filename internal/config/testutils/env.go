package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// SetupTestEnv sets up the test environment and returns a cleanup function.
// It saves the current environment, clears it, and returns a function that
// restores the original environment when called.
func SetupTestEnv(t *testing.T) func() {
	// Save current environment
	originalEnv := os.Environ()

	// Clear environment and viper config
	os.Clearenv()
	viper.Reset()

	// Get the absolute path to the testdata directory
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Set environment variables
	t.Setenv("CONFIG_FILE", configPath)
	t.Setenv("CRAWLER_SOURCE_FILE", sourcesPath)

	// Return cleanup function
	return func() {
		// Restore environment
		os.Clearenv()
		for _, e := range originalEnv {
			k, v, _ := strings.Cut(e, "=")
			t.Setenv(k, v)
		}
		viper.Reset()
	}
}

// TestSetupTestEnv verifies the test environment setup and cleanup
func TestSetupTestEnv(t *testing.T) {
	// Set a test environment variable
	t.Setenv("TEST_VAR", "test_value")

	// Setup test environment
	cleanup := SetupTestEnv(t)
	defer cleanup()

	// Verify environment is cleared
	val, exists := os.LookupEnv("TEST_VAR")
	if exists {
		t.Errorf("environment should be cleared, but TEST_VAR exists with value: %s", val)
	}

	// Verify test files are set
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		t.Error("CONFIG_FILE should be set")
	}
	sourcesFile := os.Getenv("CRAWLER_SOURCE_FILE")
	if sourcesFile == "" {
		t.Error("CRAWLER_SOURCE_FILE should be set")
	}

	// Set a new test variable
	t.Setenv("NEW_TEST_VAR", "new_value")

	// Run cleanup
	cleanup()

	// Verify original environment is restored
	val, exists = os.LookupEnv("TEST_VAR")
	if !exists || val != "test_value" {
		t.Error("original environment was not restored correctly")
	}

	// Verify new test variable is cleared
	_, exists = os.LookupEnv("NEW_TEST_VAR")
	if exists {
		t.Error("cleanup should have cleared new environment variables")
	}

	// Verify test files are cleared
	configFile = os.Getenv("CONFIG_FILE")
	if configFile != "" {
		t.Error("CONFIG_FILE should be cleared")
	}
	sourcesFile = os.Getenv("CRAWLER_SOURCE_FILE")
	if sourcesFile != "" {
		t.Error("CRAWLER_SOURCE_FILE should be cleared")
	}
}
