package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(t *testing.T) *testutils.TestSetup
		wantErrMsg string
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{}, "")
			},
			wantErrMsg: "",
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"app.environment": "invalid",
				}, "")
			},
			wantErrMsg: "invalid config: field \"app.environment\" with value invalid: invalid environment",
		},
		{
			name: "invalid log level",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"log.level": "invalid",
				}, "")
			},
			wantErrMsg: "invalid config: field \"log.level\" with value invalid: invalid log level",
		},
		{
			name: "invalid crawler max depth",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"crawler.max_depth": 0,
				}, "")
			},
			wantErrMsg: "invalid config: field \"crawler.max_depth\" with value 0: crawler max depth must be greater than 0",
		},
		{
			name: "invalid crawler parallelism",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"crawler.parallelism": 0,
				}, "")
			},
			wantErrMsg: "invalid config: field \"crawler.parallelism\" with value 0: crawler parallelism must be greater than 0",
		},
		{
			name: "server security enabled without API key",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"server.security.enabled": true,
					"server.security.api_key": "",
				}, "")
			},
			wantErrMsg: "invalid config: field \"server.security.api_key\" with value : server security is enabled but no API key is provided",
		},
		{
			name: "server security enabled with invalid API key",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, map[string]interface{}{
					"server.security.enabled": true,
					"server.security.api_key": "invalid",
				}, "")
			},
			wantErrMsg: "invalid config: field \"server.security.api_key\" with value invalid: invalid API key format",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setup := tt.setup(t)
			defer setup.Cleanup()

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestElasticsearchConfigBasicValidation(t *testing.T) {
	t.Parallel()

	// Create temporary test directory
	tmpDir := t.TempDir()
	sourcesPath := filepath.Join(tmpDir, "sources.yml")

	// Create test sources file
	sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    selectors:
      article: article
      title: h1
      content: .content
`
	err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
	require.NoError(t, err)

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
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
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
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
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
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
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
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
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
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
			},
			expectedErr: "elasticsearch API key must be in the format 'id:api_key'",
		},
		{
			name: "missing TLS certificate",
			setupEnv: func() {
				os.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				os.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				os.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				os.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				os.Setenv("GOCRAWL_ELASTICSEARCH_TLS_ENABLED", "true")
				os.Setenv("GOCRAWL_CRAWLER_SOURCE_FILE", sourcesPath)
				os.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
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

			// Create config
			cfg, err := config.New(testutils.NewTestLogger(t))
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
