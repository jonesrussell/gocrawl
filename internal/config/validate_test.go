package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testing.T)
		expectedError string
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Set required environment variables
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "",
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "invalid")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "invalid environment: invalid",
		},
		{
			name: "invalid log level",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "invalid")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "invalid log level: invalid",
		},
		{
			name: "invalid crawler max depth",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "0")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "crawler max depth must be greater than 0",
		},
		{
			name: "invalid crawler parallelism",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "0")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "crawler parallelism must be greater than 0",
		},
		{
			name: "server security enabled without API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "server security is enabled but no API key is provided",
		},
		{
			name: "server security enabled with invalid API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "invalid")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
			},
			expectedError: "server security is enabled but no API key is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := testutils.SetupTestEnv(t)
			defer cleanup()

			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.NewConfig(testutils.NewTestLogger(t))

			// Validate results
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestElasticsearchConfigBasicValidation(t *testing.T) {
	// Set up test environment
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Set base environment variables for all tests
	t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	t.Setenv("GOCRAWL_LOG_LEVEL", "info")
	t.Setenv("GOCRAWL_LOG_DEBUG", "false")
	t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
	t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
	t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
	t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")

	tests := []struct {
		name       string
		setup      func(t *testing.T)
		wantErrMsg string
	}{
		{
			name: "valid config",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
			},
			wantErrMsg: "",
		},
		{
			name: "missing addresses",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
			},
			wantErrMsg: "elasticsearch addresses cannot be empty",
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
			},
			wantErrMsg: "elasticsearch index name cannot be empty",
		},
		{
			name: "missing API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "")
			},
			wantErrMsg: "elasticsearch API key cannot be empty",
		},
		{
			name: "invalid API key format",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "test_api_key")
			},
			wantErrMsg: "elasticsearch API key must be in the format 'id:api_key'",
		},
		{
			name: "missing TLS certificate",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "https://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "id:test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("GOCRAWL_ELASTICSEARCH_TLS_CERTIFICATE", "")
			},
			wantErrMsg: "TLS certificate file is required when TLS is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.NewConfig(testutils.NewTestLogger(t))

			// Validate results
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
