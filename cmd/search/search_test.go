// Package search_test implements tests for the search command.
package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// testDeps holds all the dependencies needed for testing
type testDeps struct {
	Storage storagetypes.Interface
	Logger  types.Logger
	Config  config.Interface
	Handler *signal.SignalHandler
	Context context.Context
}

// setupTestDeps creates and configures all test dependencies
func setupTestDeps(t *testing.T) *testDeps {
	t.Helper()

	// Create mock dependencies
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := testutils.NewMockConfig()

	return &testDeps{
		Storage: mockStorage,
		Logger:  mockLogger,
		Config:  mockConfig,
		Handler: mockHandler,
		Context: t.Context(),
	}
}

// createTestApp creates a test application with the given dependencies
func createTestApp(t *testing.T, deps *testDeps) *fx.App {
	t.Helper()

	providers := []any{
		// Core dependencies
		func() storagetypes.Interface { return deps.Storage },
		func() types.Logger { return deps.Logger },
		func() config.Interface { return deps.Config },
		func() *signal.SignalHandler { return deps.Handler },
		func() context.Context { return deps.Context },
	}

	// Add command channels
	commandDone := make(chan struct{})
	providers = append(providers,
		func() chan struct{} { return commandDone },
	)

	// Create the app
	app := fx.New(
		fx.Provide(providers...),
		fx.Invoke(func(lc fx.Lifecycle, p search.Params) {
			// Add a default stop hook to clean up resources
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					// Close channels
					close(commandDone)
					return nil
				},
			})
		}),
	)

	return app
}

// runTestApp runs the test app and handles cleanup
func runTestApp(t *testing.T, app *fx.App) {
	t.Helper()

	// Start the app
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)
}

// TestCommandExecution tests the search command execution.
func TestCommandExecution(t *testing.T) {
	// Create mock dependencies
	mockStorage := testutils.NewMockStorage()
	mockLogger := logger.NewNoOp()

	// Set up mock expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]any{}, nil)
	mockStorage.On("Close").Return(nil)

	// Create test app
	app := fx.New(
		fx.NopLogger,
		fx.Supply(
			mockStorage,
			mockLogger,
		),
		fx.Invoke(func(lc fx.Lifecycle, p search.Params) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)

	// Verify mock expectations
	mockStorage.AssertExpectations(t)
}
