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
		name    string
		setup   func(*testing.T)
		wantErr bool
	}{
		{
			name: "valid_config",
			setup: func(t *testing.T) {
				// Create test config file
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  source_file: internal/config/testdata/sources.yml
logging:
  level: debug
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
			wantErr: false,
		},
		{
			name: "invalid_app_environment",
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
		{
			name: "invalid_log_level",
			setup: func(t *testing.T) {
				// Create test config file with invalid log level
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
logging:
  level: invalid
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "invalid_crawler_max_depth",
			setup: func(t *testing.T) {
				// Create test config file with invalid max depth
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
crawler:
  max_depth: 0
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "invalid_crawler_parallelism",
			setup: func(t *testing.T) {
				// Create test config file with invalid parallelism
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
crawler:
  parallelism: 0
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "server_security_enabled_without_API_key",
			setup: func(t *testing.T) {
				// Create test config file with security enabled but no API key
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
server:
  security:
    enabled: true
    api_key: ""
`
				err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")
			},
			wantErr: true,
		},
		{
			name: "server_security_enabled_with_invalid_API_key",
			setup: func(t *testing.T) {
				// Create test config file with security enabled but invalid API key
				configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
server:
  security:
    enabled: true
    api_key: "invalid"
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
