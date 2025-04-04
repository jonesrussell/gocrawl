package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestAppConfig(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testing.T)
		expectedError string
	}{
		{
			name: "valid app configuration",
			setup: func(t *testing.T) {
				// Set required environment variables
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
				t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_CRAWLER_RATE_LIMIT", "2s")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "",
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "invalid")
				t.Setenv("GOCRAWL_APP_NAME", "gocrawl-test")
				t.Setenv("GOCRAWL_APP_VERSION", "0.0.1")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_CRAWLER_RATE_LIMIT", "2s")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "invalid environment: invalid",
		},
		{
			name: "missing app configuration",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "")
				t.Setenv("GOCRAWL_APP_NAME", "")
				t.Setenv("GOCRAWL_APP_VERSION", "")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_BASE_URL", "http://test.example.com")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_CRAWLER_RATE_LIMIT", "2s")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "environment cannot be empty",
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

			// Verify app config values
			appCfg := cfg.GetAppConfig()
			require.Equal(t, "test", appCfg.Environment)
			require.Equal(t, "gocrawl-test", appCfg.Name)
			require.Equal(t, "0.0.1", appCfg.Version)
		})
	}
}

func TestAppConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T)
		wantErr bool
	}{
		{
			name: "missing_required_fields",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: test
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
				t.Setenv("GOCRAWL_CONFIG_FILE", configPath)
			},
			wantErr: true,
		},
		{
			name: "invalid_environment",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

				// Create test config file
				configContent := `
app:
  environment: invalid
  name: gocrawl
  version: 1.0.0
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
				t.Setenv("GOCRAWL_CONFIG_FILE", configPath)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test setup first to set environment variables
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
