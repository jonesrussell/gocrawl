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

	// Set environment variables in both test and actual environment
	os.Setenv("CONFIG_FILE", configPath)
	t.Setenv("CONFIG_FILE", configPath)

	os.Setenv("CRAWLER_SOURCE_FILE", sourcesPath)
	t.Setenv("CRAWLER_SOURCE_FILE", sourcesPath)

	// Debug: Print environment variables
	t.Logf("CONFIG_FILE: %s", os.Getenv("CONFIG_FILE"))
	t.Logf("CRAWLER_SOURCE_FILE: %s", os.Getenv("CRAWLER_SOURCE_FILE"))

	// Initialize Viper with the config file
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	require.NoError(t, err, "failed to read config file")

	// Debug: Print Viper config file info
	t.Logf("Viper config file: %s", viper.GetViper().ConfigFileUsed())

	// Set environment variables from config
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set environment variables and Viper values
	os.Setenv("APP_NAME", "gocrawl-test")
	t.Setenv("APP_NAME", "gocrawl-test")
	viper.Set("app.name", "gocrawl-test")

	os.Setenv("APP_VERSION", "0.0.1")
	t.Setenv("APP_VERSION", "0.0.1")
	viper.Set("app.version", "0.0.1")

	os.Setenv("APP_DEBUG", "false")
	t.Setenv("APP_DEBUG", "false")
	viper.Set("app.debug", false)

	os.Setenv("CRAWLER_BASE_URL", "http://localhost:8080")
	t.Setenv("CRAWLER_BASE_URL", "http://localhost:8080")
	viper.Set("crawler.base_url", "http://localhost:8080")

	os.Setenv("CRAWLER_MAX_DEPTH", "1")
	t.Setenv("CRAWLER_MAX_DEPTH", "1")
	viper.Set("crawler.max_depth", 1)

	os.Setenv("CRAWLER_RATE_LIMIT", "1s")
	t.Setenv("CRAWLER_RATE_LIMIT", "1s")
	viper.Set("crawler.rate_limit", "1s")

	os.Setenv("CRAWLER_PARALLELISM", "1")
	t.Setenv("CRAWLER_PARALLELISM", "1")
	viper.Set("crawler.parallelism", 1)

	os.Setenv("SERVER_ADDRESS", ":8080")
	t.Setenv("SERVER_ADDRESS", ":8080")
	viper.Set("server.address", ":8080")

	os.Setenv("SERVER_READ_TIMEOUT", "1s")
	t.Setenv("SERVER_READ_TIMEOUT", "1s")
	viper.Set("server.read_timeout", "1s")

	os.Setenv("SERVER_WRITE_TIMEOUT", "1s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "1s")
	viper.Set("server.write_timeout", "1s")

	os.Setenv("SERVER_SECURITY_ENABLED", "false")
	t.Setenv("SERVER_SECURITY_ENABLED", "false")
	viper.Set("server.security.enabled", false)

	os.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
	viper.Set("elasticsearch.addresses", []string{"http://localhost:9200"})

	os.Setenv("ELASTICSEARCH_USERNAME", "elastic")
	t.Setenv("ELASTICSEARCH_USERNAME", "elastic")
	viper.Set("elasticsearch.username", "elastic")

	os.Setenv("ELASTICSEARCH_PASSWORD", "changeme")
	t.Setenv("ELASTICSEARCH_PASSWORD", "changeme")
	viper.Set("elasticsearch.password", "changeme")

	os.Setenv("ELASTICSEARCH_INDEX_NAME", "test-index")
	t.Setenv("ELASTICSEARCH_INDEX_NAME", "test-index")
	viper.Set("elasticsearch.index_name", "test-index")

	os.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
	t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
	viper.Set("elasticsearch.retry.enabled", true)

	os.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
	t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
	viper.Set("elasticsearch.retry.initial_wait", "1s")

	os.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "5s")
	t.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "5s")
	viper.Set("elasticsearch.retry.max_wait", "5s")

	os.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "3")
	t.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "3")
	viper.Set("elasticsearch.retry.max_retries", 3)

	// Debug: Print final Viper values
	t.Logf("Final Viper values:")
	t.Logf("app.name: %s", viper.GetString("app.name"))
	t.Logf("app.version: %s", viper.GetString("app.version"))
	t.Logf("app.debug: %v", viper.GetBool("app.debug"))
	t.Logf("crawler.source_file: %s", viper.GetString("crawler.source_file"))
	t.Logf("crawler.base_url: %s", viper.GetString("crawler.base_url"))
	t.Logf("crawler.max_depth: %d", viper.GetInt("crawler.max_depth"))
	t.Logf("crawler.rate_limit: %s", viper.GetString("crawler.rate_limit"))
	t.Logf("crawler.parallelism: %d", viper.GetInt("crawler.parallelism"))
	t.Logf("server.address: %s", viper.GetString("server.address"))
	t.Logf("elasticsearch.addresses: %v", viper.GetStringSlice("elasticsearch.addresses"))
	t.Logf("elasticsearch.index_name: %s", viper.GetString("elasticsearch.index_name"))

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
