package config_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestElasticsearchConfig(t *testing.T) {
	// Get the absolute path to the testdata directory
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, *config.ElasticsearchConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, []string{"https://localhost:9200"}, cfg.Addresses)
				require.Equal(t, "test_api_key", cfg.APIKey)
				require.False(t, cfg.TLS.Enabled)
				require.Equal(t, "", cfg.TLS.CertFile)
				require.Equal(t, "", cfg.TLS.KeyFile)
			},
		},
		{
			name: "environment variable override",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_ADDRESSES", "https://override.example.com:9200")
				t.Setenv("ELASTICSEARCH_API_KEY", "override_api_key")
				t.Setenv("ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_TLS_CERTIFICATE", "override-cert.pem")
				t.Setenv("ELASTICSEARCH_TLS_KEY", "override-key.pem")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, []string{"https://override.example.com:9200"}, cfg.Addresses)
				require.Equal(t, "override_api_key", cfg.APIKey)
				require.True(t, cfg.TLS.Enabled)
				require.Equal(t, "override-cert.pem", cfg.TLS.CertFile)
				require.Equal(t, "override-key.pem", cfg.TLS.KeyFile)
			},
		},
		{
			name: "cloud configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_CLOUD_ID", "test-cloud-id")
				t.Setenv("ELASTICSEARCH_CLOUD_API_KEY", "test-cloud-api-key")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, "test-cloud-id", cfg.Cloud.ID)
				require.Equal(t, "test-cloud-api-key", cfg.Cloud.APIKey)
			},
		},
		{
			name: "retry configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "1s")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "10s")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_RETRIES", "3")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.True(t, cfg.Retry.Enabled)
				require.Equal(t, time.Second, cfg.Retry.InitialWait)
				require.Equal(t, 10*time.Second, cfg.Retry.MaxWait)
				require.Equal(t, 3, cfg.Retry.MaxRetries)
			},
		},
		{
			name: "basic auth configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_USERNAME", "test-user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "test-pass")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, "test-user", cfg.Username)
				require.Equal(t, "test-pass", cfg.Password)
			},
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
			cfg, err := config.New(testutils.NewTestLogger(t))
			require.NoError(t, err)

			// Validate results
			tt.validate(t, cfg.GetElasticsearchConfig())
		})
	}
}

func TestElasticsearchConfigValidation(t *testing.T) {
	// Get the absolute path to the testdata directory
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, error)
	}{
		{
			name: "missing API key in production",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("APP_ENV", "production")
				t.Setenv("ELASTICSEARCH_API_KEY", "")
			},
			validate: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "API key is required in production")
			},
		},
		{
			name: "invalid TLS configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_TLS_CERTIFICATE", "")
				t.Setenv("ELASTICSEARCH_TLS_KEY", "")
			},
			validate: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "certificate path cannot be empty when TLS is enabled")
			},
		},
		{
			name: "empty addresses",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_ADDRESSES", "")
			},
			validate: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "at least one Elasticsearch address must be provided")
			},
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_INDEX_NAME", "")
			},
			validate: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "index name cannot be empty")
			},
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
			cfg, err := config.New(testutils.NewTestLogger(t))

			// Validate results
			tt.validate(t, err)
			if err != nil {
				require.Nil(t, cfg)
				return
			}
			require.NotNil(t, cfg)
		})
	}
}
