package config_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// ConfigTestModule provides a test configuration module that loads the test environment file.
// This module should only be used in tests.
var ConfigTestModule = fx.Options(
	fx.Provide(
		func(t *testing.T) config.Logger { return newTestLogger(t) },
		config.New,
	),
)

func TestConfigModule(t *testing.T) {
	t.Parallel()

	// Create a test logger
	logger := newTestLogger(t)

	// Setup test config
	err := config.SetupConfig(logger, ".env.test")
	require.NoError(t, err)

	// Create test module
	module := fx.Module("test",
		fx.Provide(
			func() config.Logger { return logger },
			config.New,
		),
	)

	require.NotNil(t, module)
}
