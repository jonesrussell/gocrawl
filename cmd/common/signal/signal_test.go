// Package signal_test provides tests for the signal handler package.
package signal_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockFXApp implements the FXApp interface for testing
type mockFXApp struct {
	stopFn func(context.Context) error
}

func (m *mockFXApp) Stop(ctx context.Context) error {
	return m.stopFn(ctx)
}

// sendSignal sends a signal to the current process.
func sendSignal(t *testing.T, sig os.Signal) error {
	t.Helper()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

func TestSignalHandler(t *testing.T) {
	t.Parallel()

	t.Run("handles SIGINT", func(t *testing.T) {
		t.Parallel()
		// Create a mock logger
		mockLog := common.NewNoOpLogger()

		// Create signal handler
		handler := signalhandler.NewSignalHandler(mockLog)

		// Create a context for the test
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Set up signal handling
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT to the process
		require.NoError(t, sendSignal(t, syscall.SIGINT))

		// Wait for shutdown
		shutdownReceived := handler.Wait()
		assert.True(t, shutdownReceived)
	})

	t.Run("handles SIGTERM", func(t *testing.T) {
		t.Parallel()
		// Create a mock logger
		mockLog := common.NewNoOpLogger()

		// Create signal handler
		handler := signalhandler.NewSignalHandler(mockLog)

		// Create a context for the test
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Set up signal handling
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGTERM to the process
		require.NoError(t, sendSignal(t, syscall.SIGTERM))

		// Wait for shutdown
		shutdownReceived := handler.Wait()
		assert.True(t, shutdownReceived)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Parallel()
		// Create a mock logger
		mockLog := common.NewNoOpLogger()

		// Create signal handler
		handler := signalhandler.NewSignalHandler(mockLog)

		// Create a context for the test
		ctx, cancel := context.WithCancel(t.Context())

		// Set up signal handling
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Cancel the context
		cancel()

		// Wait for shutdown
		shutdownReceived := handler.Wait()
		assert.False(t, shutdownReceived)
	})

	t.Run("handles shutdown timeout", func(t *testing.T) {
		t.Parallel()
		// Create a mock logger
		mockLog := common.NewNoOpLogger()

		// Create signal handler
		handler := signalhandler.NewSignalHandler(mockLog)
		handler.SetShutdownTimeout(100 * time.Millisecond)

		// Create a context for the test
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Set up signal handling
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT to the process
		require.NoError(t, sendSignal(t, syscall.SIGINT))

		// Wait for shutdown
		shutdownReceived := handler.Wait()
		assert.True(t, shutdownReceived)
	})

	t.Run("coordinates with fx app shutdown", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Create a channel to track if exit was called
		exitCalled := make(chan int, 1)
		mockExit := func(code int) {
			exitCalled <- code
		}

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
		handler := signalhandler.NewSignalHandler(logger.NewNoOp())
		handler.SetExitFunc(mockExit)
		handler.SetFXApp(app)
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Start the app
		app.RequireStart()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t, syscall.SIGINT)
		}()

		// Wait for exit to be called
		select {
		case code := <-exitCalled:
			require.Equal(t, 0, code)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for exit to be called")
		}
		require.NoError(t, sigErr)

		// Verify app is stopped
		app.RequireStop()
	})

	t.Run("handles custom shutdown timeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Create a channel to track if exit was called
		exitCalled := make(chan int, 1)
		mockExit := func(code int) {
			exitCalled <- code
		}

		handler := signalhandler.NewSignalHandler(logger.NewNoOp())
		handler.SetExitFunc(mockExit)
		handler.SetShutdownTimeout(5 * time.Second)
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t, syscall.SIGINT)
		}()

		// Wait for exit to be called
		select {
		case code := <-exitCalled:
			require.Equal(t, 0, code)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for exit to be called")
		}
		require.NoError(t, sigErr)
	})

	t.Run("handles cleanup function", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Create a channel to track if exit was called
		exitCalled := make(chan int, 1)
		mockExit := func(code int) {
			exitCalled <- code
		}

		cleanupCalled := false
		handler := signalhandler.NewSignalHandler(logger.NewNoOp())
		handler.SetExitFunc(mockExit)
		handler.SetCleanup(func() {
			cleanupCalled = true
		})
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT in a goroutine
		var sigErr error
		go func() {
			time.Sleep(100 * time.Millisecond)
			sigErr = sendSignal(t, syscall.SIGINT)
		}()

		// Wait for exit to be called
		select {
		case code := <-exitCalled:
			require.Equal(t, 0, code)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for exit to be called")
		}
		require.NoError(t, sigErr)
		assert.True(t, cleanupCalled)
	})
}

func TestSignalHandler_ShutdownTimeout(t *testing.T) {
	t.Parallel()
	// Create a channel to track if exit was called
	exitCalled := make(chan int, 1)
	mockExit := func(code int) {
		exitCalled <- code
	}

	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	handler.SetExitFunc(mockExit)
	handler.SetShutdownTimeout(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Request shutdown
	handler.RequestShutdown()

	// Wait for exit to be called
	select {
	case code := <-exitCalled:
		require.Equal(t, 0, code)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for exit to be called")
	}
}

func TestSignalHandler_ShutdownTimeoutWithFX(t *testing.T) {
	t.Parallel()
	// Create a channel to track if exit was called
	exitCalled := make(chan int, 1)
	mockExit := func(code int) {
		exitCalled <- code
	}

	// Create a signal handler with a short timeout
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	handler.SetExitFunc(mockExit)
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
					shutdownComplete <- true
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

	// Request shutdown
	handler.RequestShutdown()

	// Wait for exit to be called
	select {
	case code := <-exitCalled:
		require.Equal(t, 0, code)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for exit to be called")
	}

	// Verify cleanup was called
	select {
	case <-cleanupCalled:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for cleanup to be called")
	}

	// Verify shutdown was completed
	select {
	case <-shutdownComplete:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for shutdown to complete")
	}

	// Verify app is stopped
	app.RequireStop()
}

func TestSignalHandler_ShutdownTimeoutWithError(t *testing.T) {
	t.Parallel()
	// Create a signal handler with a short timeout
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
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
					// Return a mock error that should be handled gracefully
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
		// Shutdown completed successfully
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Verify app is stopped and error is logged
	app.RequireStop()
	// Note: The error "application didn't stop cleanly: mock error" is expected
	// and is logged by fx, but doesn't affect the test result
}

func TestSignalHandler_SetLogger(t *testing.T) {
	t.Parallel()
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	newLogger := logger.NewNoOp()
	handler.SetLogger(newLogger)
	// No assertion needed - just verifying it doesn't panic
}

func TestSignalHandler_IsShuttingDown(t *testing.T) {
	t.Parallel()
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	assert.True(t, handler.IsShuttingDown(), "IsShuttingDown should return true when sigChan is initialized")

	// Test with nil sigChan
	handler = &signalhandler.SignalHandler{}
	assert.False(t, handler.IsShuttingDown(), "IsShuttingDown should return false when sigChan is nil")
}

func TestSignalHandler_ShutdownChannel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	cleanup := handler.Setup(ctx)

	// Create a channel to track when Wait returns
	done := make(chan bool)
	go func() {
		handler.Wait()
		done <- true
	}()

	// Close the shutdown channel by calling cleanup
	cleanup()

	// Wait for the handler to complete
	select {
	case <-done:
		// Success - handler completed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Handler did not complete after closing shutdown channel")
	}
}

func TestSignalHandler_ContextError(t *testing.T) {
	t.Parallel()
	// Create channels for coordination
	appReady := make(chan struct{})
	shutdownStarted := make(chan struct{})
	shutdownComplete := make(chan struct{})

	// Create a mock fx app that simulates a slow shutdown
	mockApp := &mockFXApp{
		stopFn: func(ctx context.Context) error {
			// Signal that shutdown has started
			close(shutdownStarted)
			// Simulate a slow shutdown that exceeds the context deadline
			time.Sleep(200 * time.Millisecond)
			// Signal that shutdown is complete
			close(shutdownComplete)
			return ctx.Err()
		},
	}

	// Create a signal handler with a short timeout
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	handler.SetShutdownTimeout(100 * time.Millisecond)
	handler.SetFXApp(mockApp)

	// Create a context that we'll cancel to trigger shutdown
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Setup signal handling
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Start a goroutine to wait for shutdown
	done := make(chan struct{})
	go func() {
		handler.Wait()
		close(done)
	}()

	// Start the fx app
	app := fx.New(
		fx.Supply(handler),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					close(appReady)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Wait for mock app shutdown to complete or context to be cancelled
					select {
					case <-shutdownComplete:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				},
			})
		}),
		fx.NopLogger,
	)

	// Start the app
	startCtx, startCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("Failed to start fx app: %v", err)
	}

	// Wait for app to be ready
	select {
	case <-appReady:
		// App is ready, proceed with shutdown
	case <-time.After(100 * time.Millisecond):
		t.Fatal("App did not become ready within timeout")
	}

	// Start a goroutine to request shutdown after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		handler.RequestShutdown()
	}()

	// Wait for shutdown to start
	select {
	case <-shutdownStarted:
		// Shutdown started successfully
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Shutdown did not start within timeout")
	}

	// Wait for shutdown to complete
	select {
	case <-shutdownComplete:
		// Shutdown completed successfully
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Wait for handler to complete
	select {
	case <-done:
		// Handler completed successfully
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Handler did not complete within timeout")
	}

	// Stop the fx app
	stopCtx, stopCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		// Context deadline exceeded error is expected here
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Unexpected error stopping fx app: %v", err)
		}
	}

	// Verify the final state
	if handler.GetState() != "completed" {
		t.Errorf("Expected state 'completed', got '%s'", handler.GetState())
	}
}
