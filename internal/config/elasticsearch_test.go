package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestElasticsearchConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, *config.ElasticsearchConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, []string{"https://localhost:9200"}, cfg.Addresses)
				require.Equal(t, "test_api_key", cfg.APIKey)
				require.True(t, cfg.TLS.Enabled)
				require.Equal(t, "test-cert.pem", cfg.TLS.CertFile)
				require.Equal(t, "test-key.pem", cfg.TLS.KeyFile)
			},
		},
		{
			name: "environment variable override",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: config_api_key
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
				t.Setenv("ELASTICSEARCH_API_KEY", "env_api_key")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, []string{"https://localhost:9200"}, cfg.Addresses)
				require.Equal(t, "env_api_key", cfg.APIKey)
			},
		},
		{
			name: "cloud configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
elasticsearch:
  cloud:
    id: test-cloud-id
    api_key: test-cloud-api-key
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, "test-cloud-id", cfg.Cloud.ID)
				require.Equal(t, "test-cloud-api-key", cfg.Cloud.APIKey)
			},
		},
		{
			name: "retry configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  retry:
    enabled: true
    initial_wait: 1s
    max_wait: 30s
    max_retries: 3
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.True(t, cfg.Retry.Enabled)
				require.Equal(t, time.Second, cfg.Retry.InitialWait)
				require.Equal(t, 30*time.Second, cfg.Retry.MaxWait)
				require.Equal(t, 3, cfg.Retry.MaxRetries)
			},
		},
		{
			name: "basic auth configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  username: test_user
  password: test_pass
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, "test_user", cfg.Username)
				require.Equal(t, "test_pass", cfg.Password)
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
			require.NotNil(t, cfg)

			// Validate results
			tt.validate(t, cfg.GetElasticsearchConfig())
		})
	}
}

func TestElasticsearchConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T)
		wantErr bool
	}{
		{
			name: "missing API key in production",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
app:
  environment: production
elasticsearch:
  addresses:
    - https://localhost:9200
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
				t.Setenv("APP_ENV", "production")
			},
			wantErr: true,
		},
		{
			name: "invalid TLS configuration",
			setup: func(t *testing.T) {
				// Create test config file with invalid TLS config
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  tls:
    enabled: true
    certificate: non-existent-cert.pem
    key: non-existent-key.pem
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "empty addresses",
			setup: func(t *testing.T) {
				// Create test config file with empty addresses
				configContent := `
elasticsearch:
  addresses: []
  api_key: test_api_key
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				// Create test config file with missing index name
				configContent := `
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
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
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
