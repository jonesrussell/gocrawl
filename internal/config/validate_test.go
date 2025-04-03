package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testing.T)
		expectedError string
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Set required environment variables
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "",
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "invalid")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "invalid environment: invalid",
		},
		{
			name: "invalid log level",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "invalid")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "invalid log level: invalid",
		},
		{
			name: "invalid crawler max depth",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "0")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "crawler max depth must be greater than 0",
		},
		{
			name: "invalid crawler parallelism",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "0")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "crawler parallelism must be greater than 0",
		},
		{
			name: "server security enabled without API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "server security is enabled but no API key is provided",
		},
		{
			name: "server security enabled with invalid API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "invalid")
			},
			expectedError: "invalid API key format",
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
				require.Equal(t, tt.expectedError, err.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
