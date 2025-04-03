package testutils

import (
	"os"
	"path/filepath"
	"runtime"
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
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get caller information")

	// Walk up to find the config directory
	dir := filepath.Dir(filename)
	for filepath.Base(dir) != "config" && dir != "/" {
		dir = filepath.Dir(dir)
	}
	require.NotEqual(t, "/", dir, "failed to find config directory")

	testdataDir := filepath.Join(dir, "testdata")
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Debug: Print paths
	t.Logf("Config directory: %s", dir)
	t.Logf("Testdata directory: %s", testdataDir)
	t.Logf("Config file path: %s", configPath)
	t.Logf("Sources file path: %s", sourcesPath)

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Configure Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(testdataDir)
	viper.SetEnvPrefix("GOCRAWL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	err := viper.ReadInConfig()
	require.NoError(t, err, "failed to read config file")

	// Set required environment variables
	t.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)

	// Set default environment variables for testing
	t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
	t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	t.Setenv("GOCRAWL_APP_DEBUG", "false")
	t.Setenv("GOCRAWL_LOG_LEVEL", "info")
	t.Setenv("GOCRAWL_LOG_DEBUG", "false")
	t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
	t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "test_api_key")
	t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_ENABLED", "true")
	t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
	t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_WAIT", "5s")
	t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_RETRIES", "3")

	// Return cleanup function
	return func() {
		// Restore environment
		os.Clearenv()
		for _, e := range originalEnv {
			k, v, _ := strings.Cut(e, "=")
			os.Setenv(k, v)
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
	sourcesFile := os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
	if sourcesFile == "" {
		t.Error("GOCRAWL_CRAWLER_SOURCE_FILE should be set")
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
	sourcesFile = os.Getenv("GOCRAWL_CRAWLER_SOURCE_FILE")
	if sourcesFile != "" {
		t.Error("GOCRAWL_CRAWLER_SOURCE_FILE should be cleared")
	}
}
