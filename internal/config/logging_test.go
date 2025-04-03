package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestLogConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, *config.LogConfig)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
log:
  level: debug
  debug: true
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.LogConfig) {
				require.Equal(t, "debug", cfg.Level)
				require.True(t, cfg.Debug)
			},
		},
		{
			name: "environment variable override",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
log:
  level: info
  debug: false
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
				t.Setenv("LOG_LEVEL", "warn")
				t.Setenv("LOG_DEBUG", "true")
			},
			validate: func(t *testing.T, cfg *config.LogConfig) {
				require.Equal(t, "warn", cfg.Level)
				require.True(t, cfg.Debug)
			},
		},
		{
			name: "default values",
			setup: func(t *testing.T) {
				// Create test config file with minimal config
				configContent := `{}`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			validate: func(t *testing.T, cfg *config.LogConfig) {
				require.Equal(t, "info", cfg.Level)
				require.False(t, cfg.Debug)
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
			tt.validate(t, cfg.GetLogConfig())
		})
	}
}

func TestLogConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T)
		wantErr bool
	}{
		{
			name: "invalid log level",
			setup: func(t *testing.T) {
				// Create test config file with invalid log level
				configContent := `
log:
  level: invalid
  debug: false
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
