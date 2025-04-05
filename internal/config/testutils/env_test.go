package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupTestEnvironment(t *testing.T) {
	t.Run("basic_environment_setup_and_cleanup", func(t *testing.T) {
		// Store original environment
		origEnv := make(map[string]string)
		for _, env := range os.Environ() {
			origEnv[env] = os.Getenv(env)
		}

		// Set initial environment
		t.Setenv("TEST_VAR", "test_value")

		// Setup test environment
		cleanup := SetupTestEnv(t)
		defer cleanup()

		// Verify environment is cleared
		for k := range origEnv {
			_, exists := os.LookupEnv(k)
			require.False(t, exists, "environment should be cleared, but %s exists", k)
		}

		// Verify test environment is set correctly
		for k, v := range map[string]string{
			"GOCRAWL_APP_ENVIRONMENT": "test",
			"GOCRAWL_APP_NAME":        "gocrawl-test",
			"GOCRAWL_APP_VERSION":     "0.0.1",
		} {
			actual := os.Getenv(k)
			require.Equal(t, v, actual, "environment variable %s should be set correctly", k)
		}

		// Run cleanup
		cleanup()

		// Verify cleanup restored original environment
		for k, v := range origEnv {
			actual := os.Getenv(k)
			require.Equal(t, v, actual, "cleanup should have restored original environment variable %s", k)
		}
	})
}
