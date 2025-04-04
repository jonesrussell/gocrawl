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
	// Set up test environment
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Set base environment variables for all tests
	t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
	t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
	t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
	t.Setenv("GOCRAWL_LOG_LEVEL", "info")
	t.Setenv("GOCRAWL_LOG_DEBUG", "false")

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
			if tt.wantErrMsg == "" {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}
