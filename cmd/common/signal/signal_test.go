package signal_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// setupTestHandler creates a signal handler with the given context and cleanup function.
func setupTestHandler(t *testing.T, ctx context.Context, cleanupFn func()) (*signal.SignalHandler, func()) {
	t.Helper()
	handler := signal.NewSignalHandler(logger.NewNoOp())
	if cleanupFn != nil {
		handler.SetCleanup(cleanupFn)
	}
	cleanup := handler.Setup(ctx)
	return handler, cleanup
}

// sendSignal sends a signal to the current process. If no signal is provided, SIGINT is used.
func sendSignal(t *testing.T, sig ...os.Signal) error {
	t.Helper()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	signal := syscall.SIGINT
	if len(sig) > 0 {
		if s, ok := sig[0].(syscall.Signal); ok {
			signal = s
		}
	}
	return p.Signal(signal)
}

func TestSignalHandler(t *testing.T) {
	t.Parallel()

	t.Run("handles SIGINT", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		handler, cleanup := setupTestHandler(t, ctx, nil)
		defer cleanup()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t)
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
		require.NoError(t, sigErr)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		handler, cleanup := setupTestHandler(t, ctx, nil)
		defer cleanup()

		// Cancel context in a goroutine
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		// Wait for cancellation
		assert.True(t, handler.Wait())
	})

	t.Run("handles timeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		handler, cleanup := setupTestHandler(t, ctx, nil)
		defer cleanup()

		// Create timeout context
		timeoutCtx, timeoutCancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer timeoutCancel()

		// Wait with timeout
		assert.False(t, handler.WaitWithTimeout(timeoutCtx))
	})

	t.Run("coordinates with fx app shutdown", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Create a test fx app using fxtest
		app := fxtest.New(t,
			fx.NopLogger,
			fx.Invoke(func(lc fx.Lifecycle) {
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

		// Set up the signal handler
		handler := signal.NewSignalHandler(logger.NewNoOp())
		handler.SetFXApp(app)
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Start the app
		app.RequireStart()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t)
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
		require.NoError(t, sigErr)

		// Verify app is stopped
		app.RequireStop()
	})

	t.Run("handles custom shutdown timeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		handler := signal.NewSignalHandler(logger.NewNoOp())
		handler.SetShutdownTimeout(5 * time.Second)
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t)
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
		require.NoError(t, sigErr)
	})

	t.Run("handles cleanup function", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		cleanupCalled := false
		handler, cleanup := setupTestHandler(t, ctx, func() {
			cleanupCalled = true
		})
		defer cleanup()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t)
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
		require.NoError(t, sigErr)
		assert.True(t, cleanupCalled)
	})
}

func TestSignalHandler_ShutdownTimeout(t *testing.T) {
	// Create a signal handler with a short timeout
	handler := signal.NewSignalHandler(logger.NewNoOp())
	handler.SetShutdownTimeout(100 * time.Millisecond)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Create channels to communicate test results
	cleanupCalled := make(chan bool, 1)
	shutdownComplete := make(chan bool, 1)

	// Set up cleanup function
	handler.SetCleanup(func() {
		cleanupCalled <- true
	})

	// Set up the signal handler
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Start a goroutine to wait for shutdown
	go func() {
		handler.Wait()
		shutdownComplete <- true
	}()

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for cleanup and shutdown to complete
	select {
	case <-cleanupCalled:
		// Cleanup was called
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Cleanup was not called within timeout")
	}

	select {
	case <-shutdownComplete:
		// Shutdown completed
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Shutdown did not complete within timeout")
	}
}

func TestSignalHandler_ShutdownTimeoutWithFX(t *testing.T) {
	// Create a signal handler with a short timeout
	handler := signal.NewSignalHandler(logger.NewNoOp())
	handler.SetShutdownTimeout(100 * time.Millisecond)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Create channels to communicate test results
	cleanupCalled := make(chan bool, 1)
	shutdownComplete := make(chan bool, 1)

	// Set up cleanup function
	handler.SetCleanup(func() {
		cleanupCalled <- true
	})

	// Create a test fx app using fxtest
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Invoke(func(lc fx.Lifecycle) {
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

	// Set up the signal handler
	handler.SetFXApp(app)
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Start the app
	app.RequireStart()

	// Start a goroutine to wait for shutdown
	go func() {
		handler.Wait()
		shutdownComplete <- true
	}()

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for cleanup and shutdown to complete
	select {
	case <-cleanupCalled:
		// Cleanup was called
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Cleanup was not called within timeout")
	}

	select {
	case <-shutdownComplete:
		// Shutdown completed
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Verify app is stopped
	app.RequireStop()
}

func TestSignalHandler_ShutdownTimeoutWithError(t *testing.T) {
	// Create a signal handler with a short timeout
	handler := signal.NewSignalHandler(logger.NewNoOp())
	handler.SetShutdownTimeout(100 * time.Millisecond)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Create channels to communicate test results
	cleanupCalled := make(chan bool, 1)
	shutdownComplete := make(chan bool, 1)

	// Set up cleanup function that returns an error
	handler.SetCleanup(func() {
		cleanupCalled <- true
	})

	// Create a test fx app using fxtest that returns an error on stop
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					return nil
				},
				OnStop: func(context.Context) error {
					return errors.New("mock error")
				},
			})
		}),
	)

	// Set up the signal handler
	handler.SetFXApp(app)
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Start the app
	app.RequireStart()

	// Start a goroutine to wait for shutdown
	go func() {
		handler.Wait()
		shutdownComplete <- true
	}()

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for cleanup and shutdown to complete
	select {
	case <-cleanupCalled:
		// Cleanup was called
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Cleanup was not called within timeout")
	}

	select {
	case <-shutdownComplete:
		// Shutdown completed
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Verify app is stopped
	app.RequireStop()
}
