// Package signal provides common signal handling utilities for command-line applications.
package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
)

const (
	// DefaultShutdownTimeout is the default duration to wait for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// FXApp defines the interface for fx application shutdown
type FXApp interface {
	Stop(context.Context) error
}

// ExitFunc defines a function type for system exit
type ExitFunc func(code int)

// SignalHandler manages signal handling for graceful shutdown.
type SignalHandler struct {
	// sigChan receives OS signals
	sigChan chan os.Signal
	// doneChan signals when shutdown is complete
	doneChan chan struct{}
	// cleanup is called during shutdown
	cleanup func()
	// wg tracks the signal handling goroutine
	wg sync.WaitGroup
	// fxApp is the fx application instance
	fxApp FXApp
	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout time.Duration
	// logger is the application logger
	logger common.Logger
	// once ensures we only close doneChan once
	once sync.Once
	// state tracks the current state of the handler
	state string
	// stateMu protects access to state
	stateMu sync.RWMutex
	// ctx is the context for the signal handler
	ctx context.Context
	// cancel is the cancel function for the signal handler
	cancel context.CancelFunc
	// shutdown is a channel to signal shutdown
	shutdown chan struct{}
	// exitFunc is the function used to exit the program
	exitFunc ExitFunc
	// shutdownOnce ensures we only close the shutdown channel once
	shutdownOnce sync.Once
}

// GetState returns the current state of the signal handler.
func (h *SignalHandler) GetState() string {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.state
}

// setState updates the state of the signal handler.
func (h *SignalHandler) setState(state string) {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()
	h.state = state
}

// NewSignalHandler creates a new SignalHandler instance.
func NewSignalHandler(logger common.Logger) *SignalHandler {
	return &SignalHandler{
		sigChan:         make(chan os.Signal, 1),
		doneChan:        make(chan struct{}),
		shutdownTimeout: DefaultShutdownTimeout,
		logger:          logger,
		state:           "initialized",
		shutdown:        make(chan struct{}),
		exitFunc:        nil, // Don't exit by default
	}
}

// SetShutdownTimeout sets the timeout for graceful shutdown.
func (h *SignalHandler) SetShutdownTimeout(timeout time.Duration) {
	h.shutdownTimeout = timeout
}

// SetFXApp sets the fx application instance for coordinated shutdown.
func (h *SignalHandler) SetFXApp(app FXApp) {
	h.fxApp = app
}

// SetLogger updates the logger used by the signal handler
func (h *SignalHandler) SetLogger(logger common.Logger) {
	h.logger = logger
}

// SetExitFunc sets the function used to exit the program.
// This is primarily used for testing to prevent actual program exit.
func (h *SignalHandler) SetExitFunc(exit ExitFunc) {
	h.exitFunc = exit
}

// Setup initializes signal handling for the given context.
// It returns a cleanup function that should be called when the application exits.
func (h *SignalHandler) Setup(ctx context.Context) func() {
	h.ctx, h.cancel = context.WithCancel(ctx)

	// Notify on SIGINT and SIGTERM
	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handling goroutine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		defer h.signalDone() // Ensure we signal done even if panic occurs
		h.setState("running")
		select {
		case sig := <-h.sigChan:
			// Log the received signal
			h.logger.Info("Received signal, initiating shutdown...", "signal", sig)
			h.setState("shutting_down")
			h.handleShutdown()
		case <-h.ctx.Done():
			// Context was cancelled
			h.logger.Info("Context cancelled, initiating shutdown...")
			h.setState("shutting_down")
			h.handleShutdown()
		case <-h.shutdown:
			h.logger.Info("Shutdown requested")
			h.setState("shutting_down")
			h.handleShutdown()
		}
	}()

	// Return cleanup function
	return func() {
		// Stop receiving signals
		signal.Stop(h.sigChan)
		// Close the shutdown channel to ensure the goroutine exits
		h.shutdownOnce.Do(func() {
			close(h.shutdown)
		})
		// Wait for the signal handling goroutine to finish
		h.wg.Wait()
		if h.cancel != nil {
			h.cancel()
		}
	}
}

// handleShutdown coordinates the shutdown sequence between fx and our internal cleanup.
func (h *SignalHandler) handleShutdown() {
	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownTimeout)
	defer cancel()

	// Stop the fx app if it exists
	if h.fxApp != nil {
		if err := h.fxApp.Stop(ctx); err != nil {
			h.logger.Error("Error stopping fx app", "error", err)
		}
	}

	// Call any registered cleanup functions
	if h.cleanup != nil {
		h.cleanup()
	}

	// Signal that shutdown is complete
	h.signalDone()

	// Cancel the context to ensure cleanup runs
	h.cancel()

	// Call exit function if set
	if h.exitFunc != nil {
		h.exitFunc(0)
	}
}

// SetCleanup sets a cleanup function to be called during shutdown.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.cleanup = cleanup
}

// Wait waits for a shutdown signal or context cancellation.
// It returns true if a signal was received, false if the context was cancelled.
func (h *SignalHandler) Wait() bool {
	select {
	case <-h.doneChan:
		h.setState("completed")
		if h.exitFunc != nil {
			h.exitFunc(0)
		}
		return true
	case <-h.ctx.Done():
		h.setState("completed")
		h.signalDone() // Ensure we signal done on context cancellation
		if h.exitFunc != nil {
			h.exitFunc(0)
		}
		return false
	}
}

// WaitWithTimeout waits for a shutdown signal or context cancellation with a timeout.
// It returns true if a signal was received, false if the context was cancelled or timeout occurred.
func (h *SignalHandler) WaitWithTimeout(timeoutCtx context.Context) bool {
	select {
	case <-h.doneChan:
		h.setState("completed")
		return true
	case <-timeoutCtx.Done():
		h.setState("completed")
		h.signalDone() // Ensure we signal done on timeout
		return false
	case <-h.ctx.Done():
		h.setState("completed")
		h.signalDone() // Ensure we signal done on context cancellation
		return false
	}
}

// IsShuttingDown returns true if a shutdown signal has been received.
func (h *SignalHandler) IsShuttingDown() bool {
	return h.sigChan != nil
}

// signalDone signals that shutdown is complete.
func (h *SignalHandler) signalDone() {
	h.once.Do(func() {
		close(h.doneChan)
	})
}

// RequestShutdown signals that the application should shut down.
func (h *SignalHandler) RequestShutdown() {
	h.logger.Info("Shutdown requested")
	h.shutdownOnce.Do(func() {
		close(h.shutdown)
	})
}
