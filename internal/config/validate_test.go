package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/testutils"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	// Create temporary test directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Get the absolute path to the testdata directory
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get caller information")
	dir := filepath.Dir(filename)
	sourcesPath := filepath.Join(dir, "testdata", "sources.yml")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Create test config file
	configContent := fmt.Sprintf(`
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: %s
log:
  level: debug
  debug: true
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  index_name: test-index
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
server:
  security:
    enabled: true
    api_key: id:test_api_key
`, sourcesPath)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		envValues   map[string]string
		expectedErr string
	}{
		{
			name:        "valid configuration",
			envValues:   map[string]string{},
			expectedErr: "",
		},
		{
			name: "invalid environment",
			envValues: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT": "invalid",
			},
			expectedErr: "invalid config: field \"app.environment\" with value invalid: invalid environment",
		},
		{
			name: "invalid log level",
			envValues: map[string]string{
				"GOCRAWL_LOG_LEVEL": "invalid",
			},
			expectedErr: "invalid config: field \"log.level\" with value invalid: invalid log level",
		},
		{
			name: "invalid crawler max depth",
			envValues: map[string]string{
				"GOCRAWL_CRAWLER_MAX_DEPTH": "0",
			},
			expectedErr: "invalid config: field \"crawler.max_depth\" with value 0: crawler max depth must be greater than 0",
		},
		{
			name: "invalid crawler parallelism",
			envValues: map[string]string{
				"GOCRAWL_CRAWLER_PARALLELISM": "0",
			},
			expectedErr: "invalid config: field \"crawler.parallelism\" with value 0: crawler parallelism must be greater than 0",
		},
		{
			name: "server security enabled without API key",
			envValues: map[string]string{
				"GOCRAWL_SERVER_SECURITY_ENABLED": "true",
				"GOCRAWL_SERVER_SECURITY_API_KEY": "",
			},
			expectedErr: "invalid config: field \"server.security.api_key\" with value : server security is enabled but no API key is provided",
		},
		{
			name: "server security enabled with invalid API key",
			envValues: map[string]string{
				"GOCRAWL_SERVER_SECURITY_ENABLED": "true",
				"GOCRAWL_SERVER_SECURITY_API_KEY": "invalid",
			},
			expectedErr: "invalid config: field \"server.security.api_key\" with value invalid: invalid API key format",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment with custom values
			cleanup := testutils.SetupTestEnvWithValues(t, tt.envValues)
			defer cleanup()

			// Load and validate config
			cfg, err := LoadConfig(configPath)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestElasticsearchConfigBasicValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupEnv    func()
		expectedErr string
	}{
		{
			name: "valid config",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "",
		},
		{
			name: "missing addresses",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "elasticsearch addresses cannot be empty",
		},
		{
			name: "missing index name",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "elasticsearch index name cannot be empty",
		},
		{
			name: "missing API key",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "elasticsearch API key cannot be empty",
		},
		{
			name: "invalid API key format",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "test_api_key")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "elasticsearch API key must be in the format 'id:api_key'",
		},
		{
			name: "missing TLS certificate",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "https://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				os.Setenv("GOCRAWL_ELASTICSEARCH_TLS_ENABLED", "true")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", "testdata/sources.yml")
			},
			expectedErr: "TLS certificate file is required when TLS is enabled",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			tt.setupEnv()

			// Load and validate config
			cfg, err := LoadConfig("testdata/config.yml")
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
