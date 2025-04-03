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

func TestAppConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid app configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
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

				appCfg := cfg.GetAppConfig()
				require.Equal(t, "test", appCfg.Environment)
				require.Equal(t, "gocrawl", appCfg.Name)
				require.Equal(t, "1.0.0", appCfg.Version)
				require.False(t, appCfg.Debug)
			},
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: invalid
  name: gocrawl
  version: 1.0.0
  debug: false
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
				require.Contains(t, err.Error(), "invalid environment")
			},
		},
		{
			name: "missing app configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := ``
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
				require.Contains(t, err.Error(), "app configuration is required")
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
				t.Setenv("CONFIG_FILE", configPath)
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
				t.Setenv("CONFIG_FILE", configPath)
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
