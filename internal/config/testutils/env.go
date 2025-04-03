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
	_, filename, _, ok := runtime.Caller(1)
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

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Set environment variables
	t.Setenv("CONFIG_FILE", configPath)
	t.Setenv("CRAWLER_SOURCE_FILE", sourcesPath)
	t.Setenv("APP_ENVIRONMENT", "development")
	t.Setenv("APP_NAME", "gocrawl-test")
	t.Setenv("APP_VERSION", "0.0.1")
	t.Setenv("CRAWLER_BASE_URL", "http://localhost:8080")
	t.Setenv("CRAWLER_MAX_DEPTH", "1")
	t.Setenv("CRAWLER_RATE_LIMIT", "1s")
	t.Setenv("CRAWLER_PARALLELISM", "1")
	t.Setenv("SERVER_ADDRESS", ":8080")
	t.Setenv("SERVER_READ_TIMEOUT", "1s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "1s")
	t.Setenv("SERVER_SECURITY_ENABLED", "false")
	t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	t.Setenv("ELASTICSEARCH_USERNAME", "elastic")
	t.Setenv("ELASTICSEARCH_PASSWORD", "changeme")
	t.Setenv("ELASTICSEARCH_INDEX_NAME", "test-index")
	t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
	t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
	t.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "5s")
	t.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "3")

	// Set viper values
	viper.Set("crawler.source_file", sourcesPath)
	viper.Set("app.environment", "development")
	viper.Set("app.name", "gocrawl-test")
	viper.Set("app.version", "0.0.1")
	viper.Set("crawler.base_url", "http://localhost:8080")
	viper.Set("crawler.max_depth", 1)
	viper.Set("crawler.rate_limit", "1s")
	viper.Set("crawler.parallelism", 1)
	viper.Set("server.address", ":8080")
	viper.Set("server.read_timeout", "1s")
	viper.Set("server.write_timeout", "1s")
	viper.Set("server.security.enabled", false)
	viper.Set("elasticsearch.addresses", []string{"http://localhost:9200"})
	viper.Set("elasticsearch.username", "elastic")
	viper.Set("elasticsearch.password", "changeme")
	viper.Set("elasticsearch.index_name", "test-index")
	viper.Set("elasticsearch.retry.enabled", true)
	viper.Set("elasticsearch.retry.initial_wait", "1s")
	viper.Set("elasticsearch.retry.max_wait", "5s")
	viper.Set("elasticsearch.retry.max_retries", 3)

	// Initialize Viper with config file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Set environment variables from config
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Ensure source file path is set after config is loaded
	viper.Set("crawler.source_file", sourcesPath)

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
