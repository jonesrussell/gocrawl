package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var setupMutex sync.Mutex

// TestSetup holds the test configuration and cleanup function
type TestSetup struct {
	ConfigPath  string
	SourcesPath string
	Cleanup     func()
}

// SetupTestEnvironment creates a test environment with the given configuration
func SetupTestEnvironment(t *testing.T, configContent string, sourcesContent string) *TestSetup {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	t.Helper()
	require := require.New(t)

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gocrawl-test-*")
	require.NoError(err)

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(err)

	// Create sources file
	sourcesPath := filepath.Join(tmpDir, "sources.yml")
	if sourcesContent != "" {
		err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
		require.NoError(err)
	}

	// Set environment variables
	envVars := map[string]string{
		"GOCRAWL_CONFIG_FILE":  configPath,
		"GOCRAWL_SOURCES_FILE": sourcesPath,
	}

	// Store original values
	origEnv := make(map[string]string)
	for k := range envVars {
		origEnv[k] = os.Getenv(k)
	}

	// Set new values
	for k, v := range envVars {
		err = os.Setenv(k, v)
		require.NoError(err, fmt.Sprintf("Failed to set environment variable %s", k))
	}

	cleanup := func() {
		setupMutex.Lock()
		defer setupMutex.Unlock()

		// Restore original environment variables
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}

		// Clean up temporary directory
		os.RemoveAll(tmpDir)
	}

	return &TestSetup{
		ConfigPath:  configPath,
		SourcesPath: sourcesPath,
		Cleanup:     cleanup,
	}
}
