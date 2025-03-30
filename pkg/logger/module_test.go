package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule(t *testing.T) {
	t.Parallel()

	// Create test app with logger module
	app := fxtest.New(t,
		fx.Supply(logger.Params{
			Debug:  true,
			Level:  "debug",
			AppEnv: "test",
		}),
		logger.Module,
		fx.Invoke(func(l logger.Interface) {
			// Test that we can get a logger instance
			assert.NotNil(t, l)
			// Test that we can log messages
			l.Info("test message")
		}),
	)

	// Start the app
	app.RequireStart()
	app.RequireStop()
}

func TestNewNoOp(t *testing.T) {
	t.Parallel()

	// Create a no-op logger
	l := logger.NewNoOp()

	// Test that we can call all methods without panicking
	l.Debug("test")
	l.Error("test")
	l.Info("test")
	l.Warn("test")
	l.Printf("test")
	l.Errorf("test")

	// Test that Sync returns nil
	err := l.Sync()
	require.NoError(t, err)
}
