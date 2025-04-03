package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestAppConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, *config.AppConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
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
				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
				t.Setenv("APP_ENV", "production")
				t.Setenv("APP_NAME", "gocrawl-prod")
				t.Setenv("APP_VERSION", "2.0.0")
				t.Setenv("APP_DEBUG", "true")
			},
			validate: func(t *testing.T, cfg *config.AppConfig) {
				require.Equal(t, "production", cfg.Environment)
				require.Equal(t, "gocrawl-prod", cfg.Name)
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
			require.NotNil(t, cfg)

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
				// Create test config file with missing required fields
				configContent := `
app:
  environment: test
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				// Create test config file with invalid environment
				configContent := `
app:
  environment: invalid
  name: gocrawl
  version: 1.0.0
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
