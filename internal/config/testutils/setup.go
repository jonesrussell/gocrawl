package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var setupMutex sync.Mutex

// TestSetup holds the test configuration and cleanup function
type TestSetup struct {
	ConfigPath  string
	SourcesPath string
	Cleanup     func()
}

// SetupTestEnvironment creates a test environment with the given configuration
func SetupTestEnvironment(t *testing.T, configContent string, sourcesContent string) *TestSetup {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	t.Helper()
	require := require.New(t)

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gocrawl-test-*")
	require.NoError(err)

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Create sources file
	sourcesPath := filepath.Join(tmpDir, "sources.yml")
	if sourcesContent != "" {
		err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
		require.NoError(err)
	} else {
		// Copy the default sources file
		defaultSourcesPath := filepath.Join("testdata", "configs", "sources.yml")
		sourcesContent, err := os.ReadFile(defaultSourcesPath)
		require.NoError(err)
		err = os.WriteFile(sourcesPath, sourcesContent, 0644)
		require.NoError(err)
	}

	// Set environment variables
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

	// Store original values
	origEnv := make(map[string]string)
	for k := range envVars {
		origEnv[k] = os.Getenv(k)
	}

	// Set new values
	for k, v := range envVars {
		err = os.Setenv(k, v)
		require.NoError(err, fmt.Sprintf("Failed to set environment variable %s", k))
	}

	cleanup := func() {
		setupMutex.Lock()
		defer setupMutex.Unlock()

		// Restore original environment variables
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}

		// Clean up temporary directory
		os.RemoveAll(tmpDir)
	}

	return &TestSetup{
		ConfigPath:  configPath,
		SourcesPath: sourcesPath,
		Cleanup:     cleanup,
	}
}
