package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestElasticsearchConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) string
		validate func(*testing.T, *config.ElasticsearchConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) string {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				return tmpDir
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
			setup: func(t *testing.T) string {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: config_api_key
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("ELASTICSEARCH_API_KEY", "env_api_key")
				return tmpDir
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, []string{"https://localhost:9200"}, cfg.Addresses)
				require.Equal(t, "env_api_key", cfg.APIKey)
			},
		},
		{
			name: "cloud configuration",
			setup: func(t *testing.T) string {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  cloud:
    id: test-cloud-id
    api_key: test-cloud-api-key
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				return tmpDir
			},
			validate: func(t *testing.T, cfg *config.ElasticsearchConfig) {
				require.Equal(t, "test-cloud-id", cfg.Cloud.ID)
				require.Equal(t, "test-cloud-api-key", cfg.Cloud.APIKey)
			},
		},
		{
			name: "retry configuration",
			setup: func(t *testing.T) string {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  retry:
    enabled: true
    initial_wait: 1s
    max_wait: 30s
    max_retries: 3
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				return tmpDir
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
			setup: func(t *testing.T) string {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  username: test_user
  password: test_pass
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				return tmpDir
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
			_ = tt.setup(t)

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
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: production
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			wantErr: true,
		},
		{
			name: "invalid TLS configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  tls:
    enabled: true
    certificate: ""
    key: ""
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			wantErr: true,
		},
		{
			name: "empty addresses",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses: []
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			wantErr: true,
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
crawler:
  source_file: ` + sourcesPath + `
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

				// Create test sources file
				sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
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
