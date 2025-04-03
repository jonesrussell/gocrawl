package config_test

import (
	"os"
	"path/filepath"
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
log:
  level: info
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			wantErr: false,
		},
		{
			name: "invalid_app_environment",
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
log:
  level: info
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			name: "invalid_log_level",
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
log:
  level: invalid
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			name: "invalid_crawler_max_depth",
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
log:
  level: info
crawler:
  source_file: ` + sourcesPath + `
  max_depth: -1
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			name: "invalid_crawler_parallelism",
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
log:
  level: info
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 0
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			name: "server_security_enabled_without_API_key",
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
log:
  level: info
server:
  security:
    enabled: true
    api_key: ""
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
			name: "server_security_enabled_with_invalid_API_key",
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
log:
  level: info
server:
  security:
    enabled: true
    api_key: "invalid-key"
crawler:
  source_file: ` + sourcesPath + `
  max_depth: 2
  parallelism: 4
  rate_limit: 100ms
  index_name: test-articles
  content_index_name: test-content
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
