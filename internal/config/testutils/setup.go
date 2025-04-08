package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

	// Create sources file first
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

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yml")
	if configContent != "" {
		// Update the source_file path in the config content
		if !strings.Contains(configContent, "source_file:") {
			// Parse the existing YAML to check for crawler section
			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(configContent), &config)
			require.NoError(err)

			if _, hasCrawler := config["crawler"]; !hasCrawler {
				configContent = strings.TrimSpace(configContent) + fmt.Sprintf("\ncrawler:\n  source_file: %s\n  base_url: http://test.example.com\n", sourcesPath)
			} else {
				// Unmarshal the crawler section
				var crawlerConfig map[string]interface{}
				if crawlerSection, ok := config["crawler"].(map[string]interface{}); ok {
					crawlerConfig = crawlerSection
				} else {
					crawlerConfig = make(map[string]interface{})
				}
				// Update source_file and base_url
				crawlerConfig["source_file"] = sourcesPath
				if _, hasBaseURL := crawlerConfig["base_url"]; !hasBaseURL {
					crawlerConfig["base_url"] = "http://test.example.com"
				}
				config["crawler"] = crawlerConfig

				// Marshal back to YAML
				configBytes, err := yaml.Marshal(config)
				require.NoError(err)
				configContent = string(configBytes)
			}
		} else {
			// Use yaml parsing to update the source_file path
			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(configContent), &config)
			require.NoError(err)

			if crawlerSection, ok := config["crawler"].(map[string]interface{}); ok {
				crawlerSection["source_file"] = sourcesPath
				if _, hasBaseURL := crawlerSection["base_url"]; !hasBaseURL {
					crawlerSection["base_url"] = "http://test.example.com"
				}
				config["crawler"] = crawlerSection
			}

			// Marshal back to YAML
			configBytes, err := yaml.Marshal(config)
			require.NoError(err)
			configContent = string(configBytes)
		}
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(err)
	} else {
		// Copy the default config file
		defaultConfigPath := filepath.Join("testdata", "configs", "base.yml")
		configContent, err := os.ReadFile(defaultConfigPath)
		require.NoError(err)

		// Update the source_file path using YAML parsing
		var config map[string]interface{}
		err = yaml.Unmarshal(configContent, &config)
		require.NoError(err)

		if crawlerSection, ok := config["crawler"].(map[string]interface{}); ok {
			crawlerSection["source_file"] = sourcesPath
			if _, hasBaseURL := crawlerSection["base_url"]; !hasBaseURL {
				crawlerSection["base_url"] = "http://test.example.com"
			}
			config["crawler"] = crawlerSection
		} else {
			config["crawler"] = map[string]interface{}{
				"source_file": sourcesPath,
				"base_url":    "http://test.example.com",
			}
		}

		// Marshal back to YAML
		configBytes, err := yaml.Marshal(config)
		require.NoError(err)
		err = os.WriteFile(configPath, configBytes, 0644)
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
