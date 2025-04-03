package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*testing.T)
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_config",
			setup: func(t *testing.T) {
				// Setup test environment with valid config
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				// Give time for environment variables to be set
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "invalid_app_environment",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("APP_ENVIRONMENT", "invalid")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
		{
			name: "invalid_log_level",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("LOG_LEVEL", "invalid")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
		{
			name: "invalid_crawler_max_depth",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("CRAWLER_MAX_DEPTH", "-1")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
		{
			name: "invalid_crawler_parallelism",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("CRAWLER_PARALLELISM", "0")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
		{
			name: "server_security_enabled_without_API_key",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("SERVER_SECURITY_ENABLED", "true")
				t.Setenv("SERVER_SECURITY_API_KEY", "")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
		{
			name: "server_security_enabled_with_invalid_API_key",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("SERVER_SECURITY_ENABLED", "true")
				t.Setenv("SERVER_SECURITY_API_KEY", "invalid-key")
				time.Sleep(100 * time.Millisecond)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
