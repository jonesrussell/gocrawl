// Package signal provides common signal handling utilities for command-line applications.
package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler manages signal handling for graceful shutdown.
type SignalHandler struct {
	// sigChan receives OS signals
	sigChan chan os.Signal
	// doneChan signals when shutdown is complete
	doneChan chan struct{}
	// cleanup is called during shutdown
	cleanup func()
}

// NewSignalHandler creates a new SignalHandler instance.
func NewSignalHandler() *SignalHandler {
	return &SignalHandler{
		sigChan:  make(chan os.Signal, 1),
		doneChan: make(chan struct{}),
	}
}

// Setup initializes signal handling for the given context.
// It returns a cleanup function that should be called when the application exits.
func (h *SignalHandler) Setup(ctx context.Context) func() {
	// Notify on SIGINT and SIGTERM
	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handling goroutine
	go func() {
		select {
		case sig := <-h.sigChan:
			// Log the received signal
			os.Stderr.WriteString("\nReceived signal " + sig.String() + ", initiating shutdown...\n")
		case <-ctx.Done():
			// Context was cancelled
			os.Stderr.WriteString("\nContext cancelled, initiating shutdown...\n")
		}
		close(h.doneChan)
	}()

	// Return cleanup function
	return func() {
		signal.Stop(h.sigChan)
		if h.cleanup != nil {
			h.cleanup()
		}
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
