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

func TestPriorityConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid priority configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
priority:
  default: 1
  rules:
    - pattern: ".*\\.html$"
      priority: 2
    - pattern: ".*/article/.*"
      priority: 3
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

				priorityCfg := cfg.GetPriorityConfig()
				require.Equal(t, 1, priorityCfg.Default)
				require.Len(t, priorityCfg.Rules, 2)

				// Verify first rule
				require.Equal(t, ".*\\.html$", priorityCfg.Rules[0].Pattern)
				require.Equal(t, 2, priorityCfg.Rules[0].Priority)

				// Verify second rule
				require.Equal(t, ".*/article/.*", priorityCfg.Rules[1].Pattern)
				require.Equal(t, 3, priorityCfg.Rules[1].Priority)
			},
		},
		{
			name: "invalid priority pattern",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
priority:
  default: 1
  rules:
    - pattern: "["
      priority: 2
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
				require.Contains(t, err.Error(), "invalid priority pattern")
			},
		},
		{
			name: "invalid priority value",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")

				// Create test config file
				configContent := `
app:
  environment: test
priority:
  default: -1
  rules:
    - pattern: ".*\\.html$"
      priority: 2
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
				require.Contains(t, err.Error(), "priority must be greater than or equal to 0")
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
