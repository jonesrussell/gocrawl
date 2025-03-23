// Package signal provides common signal handling utilities for command-line applications.
package signal

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

const (
	// DefaultShutdownTimeout is the default duration to wait for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// FXApp defines the interface for fx application shutdown
type FXApp interface {
	Stop(context.Context) error
}

// SignalHandler manages signal handling for graceful shutdown.
type SignalHandler struct {
	// sigChan receives OS signals
	sigChan chan os.Signal
	// doneChan signals when shutdown is complete
	doneChan chan struct{}
	// cleanup is called during shutdown
	cleanup func()
	// isServer indicates if this is a server mode handler
	isServer bool
	// wg tracks the signal handling goroutine
	wg sync.WaitGroup
	// fxApp is the fx application instance
	fxApp FXApp
	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout time.Duration
	// fxDoneChan signals when fx app shutdown is complete
	fxDoneChan chan struct{}
	// logger is the application logger
	logger logger.Interface
}

// NewSignalHandler creates a new SignalHandler instance.
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		sigChan:         make(chan os.Signal, 1),
		doneChan:        make(chan struct{}),
		fxDoneChan:      make(chan struct{}),
		shutdownTimeout: DefaultShutdownTimeout,
		logger:          logger,
	}
}

// NewServerSignalHandler creates a new SignalHandler instance for server mode.
func NewServerSignalHandler() *SignalHandler {
	return &SignalHandler{
		sigChan:         make(chan os.Signal, 1),
		doneChan:        make(chan struct{}),
		isServer:        true,
		shutdownTimeout: DefaultShutdownTimeout,
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
func (h *SignalHandler) SetLogger(logger logger.Interface) {
	h.logger = logger
}

// Setup initializes signal handling for the given context.
// It returns a cleanup function that should be called when the application exits.
func (h *SignalHandler) Setup(ctx context.Context) func() {
	// Notify on SIGINT and SIGTERM
	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handling goroutine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		select {
		case sig := <-h.sigChan:
			// Log the received signal
			h.logger.Info("Received signal, initiating shutdown...", "signal", sig)
			h.handleShutdown(ctx)
		case <-ctx.Done():
			// Context was cancelled
			h.logger.Info("Context cancelled, initiating shutdown...")
			h.handleShutdown(ctx)
		}
		close(h.doneChan)
	}()

	// Return cleanup function
	return func() {
		// Stop receiving signals
		signal.Stop(h.sigChan)
		// Wait for the signal handling goroutine to finish
		h.wg.Wait()
	}
}

// handleShutdown coordinates the shutdown sequence between fx and our internal cleanup.
func (h *SignalHandler) handleShutdown(ctx context.Context) {
	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, h.shutdownTimeout)
	defer cancel()

	// If we have an fx app, stop it first
	if h.fxApp != nil {
		if err := h.fxApp.Stop(shutdownCtx); err != nil {
			if !isContextError(err) {
				h.logger.Error("Error during fx shutdown", "error", err)
			}
		}
	}

	// Call any registered cleanup functions
	if h.cleanup != nil {
		h.cleanup()
	}
}

// SetCleanup sets a cleanup function to be called during shutdown.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.cleanup = cleanup
}

// Wait waits for a shutdown signal or context cancellation.
// It returns true if a signal was received, false if the context was cancelled.
func (h *SignalHandler) Wait() bool {
	<-h.doneChan
	return true
}

// WaitWithTimeout waits for a shutdown signal or context cancellation with a timeout.
// It returns true if a signal was received, false if the context was cancelled or timeout occurred.
func (h *SignalHandler) WaitWithTimeout(timeoutCtx context.Context) bool {
	select {
	case <-h.doneChan:
		return true
	case <-timeoutCtx.Done():
		return false
	}
}

// IsShuttingDown returns true if a shutdown signal has been received.
func (h *SignalHandler) IsShuttingDown() bool {
	select {
	case <-h.doneChan:
		return true
	default:
		return false
	}
}

// isContextError checks if an error is a context-related error.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
