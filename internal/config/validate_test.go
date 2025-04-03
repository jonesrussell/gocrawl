package config_test

import (
	"os"
	"testing"

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
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr: false,
		},
		{
			name: "invalid_app_environment",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "invalid")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr:     true,
			errContains: "invalid environment",
		},
		{
			name: "invalid_log_level",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "invalid")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid_crawler_max_depth",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "-1")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr:     true,
			errContains: "max depth must be greater than 0",
		},
		{
			name: "invalid_crawler_parallelism",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "0")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr:     true,
			errContains: "parallelism must be greater than 0",
		},
		{
			name: "server_security_enabled_without_API_key",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			wantErr:     true,
			errContains: "API key is required when security is enabled",
		},
		{
			name: "server_security_enabled_with_invalid_API_key",
			setup: func(t *testing.T) {
				cleanup := testutils.SetupTestEnv(t)
				defer cleanup()
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "development")
				t.Setenv("GOCRAWL_LOG_LEVEL", "info")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "invalid-key")
			},
			wantErr:     true,
			errContains: "invalid API key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test setup
			tt.setup(t)

			// Debug: Print environment variables
			t.Logf("Environment variables after setup:")
			for _, env := range os.Environ() {
				t.Logf("  %s", env)
			}

			// Create config
			cfg, err := config.NewConfig(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Debug: Print config values
			t.Logf("Config values:")
			t.Logf("  App Environment: %s", cfg.GetAppConfig().Environment)
			t.Logf("  Log Level: %s", cfg.GetLogConfig().Level)
			t.Logf("  Crawler Max Depth: %d", cfg.GetCrawlerConfig().MaxDepth)
			t.Logf("  Crawler Parallelism: %d", cfg.GetCrawlerConfig().Parallelism)
			t.Logf("  Server Security Enabled: %v", cfg.GetServerConfig().Security.Enabled)
			t.Logf("  Server Security API Key: %s", cfg.GetServerConfig().Security.APIKey)
		})
	}
}
