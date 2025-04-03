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
		validate func(*testing.T, *config.AppConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			validate: func(t *testing.T, cfg *config.AppConfig) {
				require.Equal(t, "test", cfg.Environment)
				require.Equal(t, "gocrawl", cfg.Name)
				require.Equal(t, "1.0.0", cfg.Version)
				require.False(t, cfg.Debug)
			},
		},
		{
			name: "environment variable override",
			setup: func(t *testing.T) {
				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
				t.Setenv("APP_ENVIRONMENT", "development")
				t.Setenv("APP_NAME", "test-app")
				t.Setenv("APP_VERSION", "2.0.0")
				t.Setenv("APP_DEBUG", "true")
			},
			validate: func(t *testing.T, cfg *config.AppConfig) {
				require.Equal(t, "development", cfg.Environment)
				require.Equal(t, "test-app", cfg.Name)
				require.Equal(t, "2.0.0", cfg.Version)
				require.True(t, cfg.Debug)
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
			tt.validate(t, cfg.GetAppConfig())
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
			name: "missing required fields",
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
			name: "invalid environment",
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
