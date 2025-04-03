package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestElasticsearchConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid elasticsearch configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  index_name: test-index
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Configure Viper
				viper.SetConfigFile(configPath)
				viper.SetConfigType("yaml")
				err = viper.ReadInConfig()
				require.NoError(t, err)
			},
			validate: func(t *testing.T, cfg config.Interface, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)

				esCfg := cfg.GetElasticsearchConfig()
				require.Equal(t, []string{"https://localhost:9200"}, esCfg.Addresses)
				require.Equal(t, "test_api_key", esCfg.APIKey)
				require.Equal(t, "test-index", esCfg.IndexName)
				require.True(t, esCfg.TLS.Enabled)
				require.Equal(t, "test-cert.pem", esCfg.TLS.CertFile)
				require.Equal(t, "test-key.pem", esCfg.TLS.KeyFile)
			},
		},
		{
			name: "missing elasticsearch configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Configure Viper
				viper.SetConfigFile(configPath)
				viper.SetConfigType("yaml")
				err = viper.ReadInConfig()
				require.NoError(t, err)
			},
			validate: func(t *testing.T, cfg config.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "elasticsearch configuration is required")
			},
		},
		{
			name: "invalid elasticsearch address",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
elasticsearch:
  addresses:
    - invalid-url
  api_key: test_api_key
  index_name: test-index
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Configure Viper
				viper.SetConfigFile(configPath)
				viper.SetConfigType("yaml")
				err = viper.ReadInConfig()
				require.NoError(t, err)
			},
			validate: func(t *testing.T, cfg config.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid elasticsearch address")
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
			cfg, err := config.NewConfig(testutils.NewTestLogger(t))

			// Validate results
			tt.validate(t, cfg, err)
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
			name: "invalid retry initial wait",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for invalid retry initial wait test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_RETRY_ENABLED: %s", os.Getenv("ELASTICSEARCH_RETRY_ENABLED"))
				t.Logf("ELASTICSEARCH_RETRY_INITIAL_WAIT: %s", os.Getenv("ELASTICSEARCH_RETRY_INITIAL_WAIT"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
				t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "500ms")
			},
			expectedErr: "initial wait must be at least 1 second",
		},
		{
			name: "invalid retry max wait",
			setup: func(t *testing.T) {
				// Debug: Print environment variables
				t.Logf("Setting up environment variables for invalid retry max wait test")
				t.Logf("ELASTICSEARCH_ADDRESSES: %s", os.Getenv("ELASTICSEARCH_ADDRESSES"))
				t.Logf("ELASTICSEARCH_USERNAME: %s", os.Getenv("ELASTICSEARCH_USERNAME"))
				t.Logf("ELASTICSEARCH_PASSWORD: %s", os.Getenv("ELASTICSEARCH_PASSWORD"))
				t.Logf("ELASTICSEARCH_RETRY_ENABLED: %s", os.Getenv("ELASTICSEARCH_RETRY_ENABLED"))
				t.Logf("ELASTICSEARCH_RETRY_INITIAL_WAIT: %s", os.Getenv("ELASTICSEARCH_RETRY_INITIAL_WAIT"))
				t.Logf("ELASTICSEARCH_RETRY_MAX_WAIT: %s", os.Getenv("ELASTICSEARCH_RETRY_MAX_WAIT"))

				t.Setenv("ELASTICSEARCH_ADDRESSES", "http://localhost:9200")
				t.Setenv("ELASTICSEARCH_USERNAME", "user")
				t.Setenv("ELASTICSEARCH_PASSWORD", "pass")
				t.Setenv("ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("ELASTICSEARCH_RETRY_INITIAL_WAIT", "2s")
				t.Setenv("ELASTICSEARCH_RETRY_MAX_WAIT", "1s")
			},
			expectedErr: "max wait must be greater than or equal to initial wait",
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
