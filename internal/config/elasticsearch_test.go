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

	// Print initial environment state
	t.Log("Starting TestElasticsearchConfigValidation")
	t.Log("Initial environment variables:")
	for _, env := range os.Environ() {
		t.Logf("  %s", env)
	}

	// Set log level first to ensure it's preserved
	t.Setenv("GOCRAWL_LOG_LEVEL", "info")
	t.Setenv("GOCRAWL_LOG_DEBUG", "false")

	// Print environment after SetupTestEnv
	t.Log("Environment after SetupTestEnv:")
	for _, env := range os.Environ() {
		t.Logf("  %s", env)
	}

	// Debug: Print Viper configuration
	t.Log("Viper configuration:")
	t.Logf("  Config file: %s", viper.GetViper().ConfigFileUsed())
	t.Logf("  Env prefix: %s", viper.GetViper().GetEnvPrefix())
	t.Logf("  Automatic env: %v", viper.GetViper().IsSet("automatic_env"))
	t.Logf("  All settings:")
	for _, key := range viper.AllKeys() {
		t.Logf("    %s: %v", key, viper.Get(key))
	}

	tests := []struct {
		name       string
		setup      func(t *testing.T)
		wantErrMsg string
	}{
		{
			name: "missing addresses",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for missing addresses test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_CLOUD_ID: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_CLOUD_ID"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_CLOUD_ID", "")
				t.Log("Environment after setting up missing addresses test:")
				for _, env := range os.Environ() {
					t.Logf("  %s", env)
				}
			},
			wantErrMsg: "at least one Elasticsearch address must be provided",
		},
		{
			name: "invalid address",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for invalid address test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "not-a-url")
				t.Log("Environment after setting up invalid address test:")
				for _, env := range os.Environ() {
					t.Logf("  %s", env)
				}
			},
			wantErrMsg: "invalid Elasticsearch address",
		},
		{
			name: "missing credentials",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for missing credentials test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Logf("GOCRAWL_ELASTICSEARCH_API_KEY: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_API_KEY"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_USERNAME", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_PASSWORD", "")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "")
				t.Log("Environment after setting up missing credentials test:")
				for _, env := range os.Environ() {
					t.Logf("  %s", env)
				}
			},
			wantErrMsg: "either username/password or api_key must be provided",
		},
		{
			name: "invalid retry values",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for invalid retry values test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_RETRIES", "-1")
			},
			wantErrMsg: "max retries must be greater than or equal to 0",
		},
		{
			name: "invalid retry initial wait",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for invalid retry initial wait test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT", "500ms")
			},
			wantErrMsg: "initial wait must be at least 1 second",
		},
		{
			name: "invalid retry max wait",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for invalid retry max wait test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_ENABLED", "true")
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_INITIAL_WAIT", "2s")
				t.Setenv("GOCRAWL_ELASTICSEARCH_RETRY_MAX_WAIT", "1s")
			},
			wantErrMsg: "max wait must be greater than initial wait",
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for missing index name test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "")
			},
			wantErrMsg: "index name cannot be empty",
		},
		{
			name: "missing TLS certificate",
			setup: func(t *testing.T) {
				t.Log("Setting up environment variables for missing TLS certificate test")
				t.Logf("GOCRAWL_ELASTICSEARCH_ADDRESSES: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_ADDRESSES"))
				t.Logf("GOCRAWL_ELASTICSEARCH_USERNAME: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_USERNAME"))
				t.Logf("GOCRAWL_ELASTICSEARCH_PASSWORD: %s", os.Getenv("GOCRAWL_ELASTICSEARCH_PASSWORD"))
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_ELASTICSEARCH_ADDRESSES", "https://localhost:9200")
				t.Setenv("GOCRAWL_ELASTICSEARCH_API_KEY", "test_api_key")
				t.Setenv("GOCRAWL_ELASTICSEARCH_INDEX_NAME", "test-index")
				t.Setenv("GOCRAWL_ELASTICSEARCH_TLS_ENABLED", "true")
				t.Setenv("GOCRAWL_ELASTICSEARCH_TLS_CERTIFICATE", "")
			},
			wantErrMsg: "certificate path cannot be empty when TLS is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tt.name)

			// Debug: Print environment before test setup
			t.Logf("Environment before test setup:")
			for _, env := range os.Environ() {
				t.Logf("  %s", env)
			}

			tt.setup(t)

			// Debug: Print environment after test setup
			t.Logf("Environment after test setup:")
			for _, env := range os.Environ() {
				t.Logf("  %s", env)
			}

			// Debug: Print viper configuration
			t.Logf("Viper configuration:")
			t.Logf("  Config file: %s", viper.GetViper().ConfigFileUsed())
			t.Logf("  Elasticsearch addresses: %v", viper.GetStringSlice("elasticsearch.addresses"))
			t.Logf("  Elasticsearch index name: %s", viper.GetString("elasticsearch.index_name"))
			t.Logf("  Elasticsearch API key: %s", viper.GetString("elasticsearch.api_key"))

			_, err := config.New(testutils.NewTestLogger(t))
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrMsg)

			// Debug: Print test result
			t.Logf("Test case %s completed successfully", tt.name)
		})
	}
}
