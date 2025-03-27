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
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		sigChan:         make(chan os.Signal, 1),
		doneChan:        make(chan struct{}),
		fxDoneChan:      make(chan struct{}),
		shutdownTimeout: DefaultShutdownTimeout,
		logger:          logger,
		state:           "initialized",
		shutdown:        make(chan struct{}),
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
	h.ctx, h.cancel = context.WithCancel(ctx)

	// Notify on SIGINT and SIGTERM
	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handling goroutine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.setState("running")
		select {
		case sig := <-h.sigChan:
			// Log the received signal
			h.logger.Info("Received signal, initiating shutdown...", "signal", sig)
			h.setState("shutting_down")
			h.handleShutdown(ctx)
		case <-h.ctx.Done():
			// Context was cancelled
			h.logger.Info("Context cancelled, initiating shutdown...")
			h.setState("shutting_down")
			h.handleShutdown(ctx)
		case <-h.shutdown:
			h.logger.Info("Shutdown requested")
		}
		h.signalDone()
	}()

	// Return cleanup function
	return func() {
		// Stop receiving signals
		signal.Stop(h.sigChan)
		// Wait for the signal handling goroutine to finish
		h.wg.Wait()
		h.setState("cleaned_up")
		if h.cancel != nil {
			h.cancel()
		}
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
	h.setState("completed")
	return true
}

// WaitWithTimeout waits for a shutdown signal or context cancellation with a timeout.
// It returns true if a signal was received, false if the context was cancelled or timeout occurred.
func (h *SignalHandler) WaitWithTimeout(timeoutCtx context.Context) bool {
	select {
	case <-h.doneChan:
		h.setState("completed")
		return true
	case <-timeoutCtx.Done():
		h.setState("timeout")
		return false
	}
}

// IsShuttingDown returns true if a shutdown signal has been received.
func (h *SignalHandler) IsShuttingDown() bool {
	return h.sigChan != nil
}

// Complete signals that the operation is complete and should exit.
func (h *SignalHandler) Complete() {
	h.setState("completing")
	h.signalDone()
}

// signalDone ensures the doneChan is closed exactly once.
func (h *SignalHandler) signalDone() {
	h.once.Do(func() {
		if h.doneChan != nil {
			close(h.doneChan)
		}
	})
}

// isContextError checks if an error is a context-related error.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// RequestShutdown signals that the application should shut down.
func (h *SignalHandler) RequestShutdown() {
	select {
	case h.shutdown <- struct{}{}:
	default:
		// Channel is already closed or full, ignore
	}
}
