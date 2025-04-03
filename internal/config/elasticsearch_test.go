package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchConfig(t *testing.T) {
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	logger := testutils.NewTestLogger(t)

	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, *config.ElasticsearchConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for valid configuration")
				t.Logf("CONFIG_FILE: %s", os.Getenv("CONFIG_FILE"))
				t.Logf("CRAWLER_SOURCE_FILE: %s", os.Getenv("CRAWLER_SOURCE_FILE"))
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_INDEX_NAME: %s", os.Getenv("ELASTICSEARCH_INDEX_NAME"))
				t.Logf("ELASTICSEARCH_TLS_ENABLED: %s", os.Getenv("ELASTICSEARCH_TLS_ENABLED"))
				t.Logf("ELASTICSEARCH_TLS_CERTIFICATE: %s", os.Getenv("ELASTICSEARCH_TLS_CERTIFICATE"))
				t.Logf("ELASTICSEARCH_TLS_KEY: %s", os.Getenv("ELASTICSEARCH_TLS_KEY"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "elastic")
				t.Setenv("ELASTICSEARCH_PASSWORD", "changeme")
				t.Setenv("ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_TLS_CERTIFICATE", "/path/to/cert")
				t.Setenv("ELASTICSEARCH_TLS_KEY", "/path/to/key")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				// Debug: Print config values
				t.Logf("Validating Elasticsearch config:")
				t.Logf("Addresses: %v", cfg.Addresses)
				t.Logf("Username: %s", cfg.Username)
				t.Logf("Password: %s", cfg.Password)
				t.Logf("IndexName: %s", cfg.IndexName)
				t.Logf("TLS.Enabled: %v", cfg.TLS.Enabled)
				t.Logf("TLS.CertFile: %s", cfg.TLS.CertFile)
				t.Logf("TLS.KeyFile: %s", cfg.TLS.KeyFile)
				t.Logf("Retry.Enabled: %v", cfg.Retry.Enabled)
				t.Logf("Retry.InitialWait: %v", cfg.Retry.InitialWait)
				t.Logf("Retry.MaxWait: %v", cfg.Retry.MaxWait)
				t.Logf("Retry.MaxRetries: %d", cfg.Retry.MaxRetries)

				require.Equal(t, []string{"http://localhost:9200"}, cfg.Addresses)
				require.Equal(t, "elastic", cfg.Username)
				require.Equal(t, "changeme", cfg.Password)
				require.Equal(t, "test-index", cfg.IndexName)
				require.True(t, cfg.TLS.Enabled)
				require.Equal(t, "/path/to/cert", cfg.TLS.CertFile)
				require.Equal(t, "/path/to/key", cfg.TLS.KeyFile)
			},
		},
		{
			name: "cloud configuration",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for cloud configuration")
				t.Logf("ELASTICSEARCH_CLOUD_ID: %s", os.Getenv("ELASTICSEARCH_CLOUD_ID"))
				t.Logf("ELASTICSEARCH_API_KEY: %s", os.Getenv("ELASTICSEARCH_API_KEY"))

				t.Setenv("ELASTICSEARCH_CLOUD_ID", "test:dGVzdC5lbGFzdGljc2VhcmNoLm5ldA==")
				t.Setenv("ELASTICSEARCH_API_KEY", "test-api-key")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				// Debug: Print config values
				t.Logf("Validating cloud config:")
				t.Logf("Cloud.ID: %s", cfg.Cloud.ID)
				t.Logf("APIKey: %s", cfg.APIKey)

				require.Equal(t, "test:dGVzdC5lbGFzdGljc2VhcmNoLm5ldA==", cfg.Cloud.ID)
				require.Equal(t, "test-api-key", cfg.APIKey)
			},
		},
		{
			name: "retry configuration",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for retry configuration")
				t.Logf("ELASTICSEARCH_RETRY_ENABLED: %s", os.Getenv("ELASTICSEARCH_RETRY_ENABLED"))
				t.Logf("ELASTICSEARCH_RETRY_INITIAL_WAIT: %s", os.Getenv("ELASTICSEARCH_RETRY_INITIAL_WAIT"))
				t.Logf("ELASTICSEARCH_RETRY_MAX_WAIT: %s", os.Getenv("ELASTICSEARCH_RETRY_MAX_WAIT"))
				t.Logf("ELASTICSEARCH_RETRY_MAX_RETRIES: %s", os.Getenv("ELASTICSEARCH_RETRY_MAX_RETRIES"))

				t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "5s")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "3")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				// Debug: Print config values
				t.Logf("Validating retry config:")
				t.Logf("Retry.Enabled: %v", cfg.Retry.Enabled)
				t.Logf("Retry.InitialWait: %v", cfg.Retry.InitialWait)
				t.Logf("Retry.MaxWait: %v", cfg.Retry.MaxWait)
				t.Logf("Retry.MaxRetries: %d", cfg.Retry.MaxRetries)

				require.True(t, cfg.Retry.Enabled)
				require.Equal(t, time.Second, cfg.Retry.InitialWait)
				require.Equal(t, 5*time.Second, cfg.Retry.MaxWait)
				require.Equal(t, 3, cfg.Retry.MaxRetries)
			},
		},
		{
			name: "basic auth configuration",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for basic auth configuration")
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))

				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				// Debug: Print config values
				t.Logf("Validating basic auth config:")
				t.Logf("Username: %s", cfg.Username)
				t.Logf("Password: %s", cfg.Password)

				require.Equal(t, "user", cfg.Username)
				require.Equal(t, "pass", cfg.Password)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			cfg, err := config.New(logger)
			require.NoError(t, err)
			tt.validate(t, cfg.GetElasticsearchConfig())
		})
	}
}

func TestElasticsearchConfigValidation(t *testing.T) {
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	logger := testutils.NewTestLogger(t)

	tests := []struct {
		name        string
		setup       func(*testing.T)
		expectedErr string
	}{
		{
			name: "missing addresses",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for missing addresses test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_CLOUD_ID: %s", os.Getenv("ELASTICSEARCH_CLOUD_ID"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "")
				t.Setenv("ELASTICSEARCH_CLOUD_ID", "")
			},
			expectedErr: "at least one Elasticsearch address must be provided",
		},
		{
			name: "invalid address",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for invalid address test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "not-a-url")
			},
			expectedErr: "invalid Elasticsearch address",
		},
		{
			name: "missing credentials",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for missing credentials test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_API_KEY: %s", os.Getenv("ELASTICSEARCH_API_KEY"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "")
				t.Setenv("ELASTICSEARCH_PASSWORD", "")
				t.Setenv("ELASTICSEARCH_API_KEY", "")
			},
			expectedErr: "either username/password or api_key must be provided",
		},
		{
			name: "invalid retry values",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for invalid retry values test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_RETRY_ENABLED: %s", os.Getenv("ELASTICSEARCH_RETRY_ENABLED"))
				t.Logf("ELASTICSEARCH_RETRY_MAX_RETRIES: %s", os.Getenv("ELASTICSEARCH_RETRY_MAX_RETRIES"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
				t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "-1")
			},
			expectedErr: "retry values must be non-negative",
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for missing index name test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_INDEX_NAME: %s", os.Getenv("ELASTICSEARCH_INDEX_NAME"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
				t.Setenv("ELASTICSEARCH_INDEX_NAME", "")
			},
			expectedErr: "index name cannot be empty",
		},
		{
			name: "missing TLS certificate",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for missing TLS certificate test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_TLS_ENABLED: %s", os.Getenv("ELASTICSEARCH_TLS_ENABLED"))
				t.Logf("ELASTICSEARCH_TLS_CERTIFICATE: %s", os.Getenv("ELASTICSEARCH_TLS_CERTIFICATE"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
				t.Setenv("ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_TLS_CERTIFICATE", "")
			},
			expectedErr: "certificate path cannot be empty when TLS is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			_, err := config.New(logger)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
