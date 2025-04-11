// Package signal provides signal handling functionality.
package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// shutdownState represents the current state of the signal handler
type shutdownState int

const (
	stateRunning shutdownState = iota
	stateShuttingDown
	stateShutdownComplete
)

// Interface defines the signal handler interface.
type Interface interface {
	Setup(ctx context.Context) func()
	SetFXApp(app interface{})
	RequestShutdown()
	Wait() error
	AddResource(closer func() error)
	GetState() string
	SetLogger(logger logger.Interface)
	SetCleanup(cleanup func())
	IsShuttingDown() bool
}

// SignalHandler handles OS signals and graceful shutdown.
type SignalHandler struct {
	logger          logger.Interface
	done            chan struct{}
	mu              sync.Mutex
	app             any // Can be *fx.App or func() error
	state           shutdownState
	stateMu         sync.RWMutex
	resources       []func() error
	resourcesMu     sync.Mutex
	shutdownTimeout time.Duration
	cleanup         func()
	shutdownError   error
	signals         []os.Signal
}

// NewSignalHandler creates a new signal handler.
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		logger:          logger,
		done:            make(chan struct{}),
		state:           stateRunning,
		shutdownTimeout: DefaultShutdownTimeout,
	}
}

// Setup sets up signal handling.
func (h *SignalHandler) Setup(ctx context.Context) func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			h.logger.Info("Received signal, initiating shutdown", "signal", sig)
			h.RequestShutdown()
		case <-ctx.Done():
			h.logger.Info("Context cancelled, initiating shutdown")
			h.RequestShutdown()
		}
	}()

	return func() {
		signal.Stop(sigChan)
		close(sigChan)
	}
}

// Wait blocks until shutdown is complete and returns any error that occurred during shutdown
func (h *SignalHandler) Wait() error {
	<-h.done
	return h.shutdownError
}

// RequestShutdown initiates graceful shutdown
func (h *SignalHandler) RequestShutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	select {
	case <-h.done:
		// Already closed
	default:
		h.shutdown() // Call shutdown before closing done channel
		close(h.done)
	}
}

// SetFXApp sets the FX app instance.
func (h *SignalHandler) SetFXApp(app any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.app = app
}

// AddResource adds a resource that needs to be closed during shutdown.
func (h *SignalHandler) AddResource(closer func() error) {
	h.resourcesMu.Lock()
	defer h.resourcesMu.Unlock()
	h.resources = append(h.resources, closer)
}

// GetState returns the current state of the signal handler.
func (h *SignalHandler) GetState() string {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()

	switch h.state {
	case stateRunning:
		return "running"
	case stateShuttingDown:
		return "shutting down"
	case stateShutdownComplete:
		return "completed"
	default:
		return "unknown"
	}
}

// SetLogger sets the logger for the signal handler.
func (h *SignalHandler) SetLogger(logger logger.Interface) {
	h.logger = logger
}

// SetCleanup sets the cleanup function to be called during shutdown.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.cleanup = cleanup
}

// IsShuttingDown returns whether the signal handler is in the process of shutting down.
func (h *SignalHandler) IsShuttingDown() bool {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.state == stateShuttingDown
}

// SetShutdownTimeout sets the shutdown timeout.
func (h *SignalHandler) SetShutdownTimeout(timeout time.Duration) {
	h.shutdownTimeout = timeout
}

// shutdown performs the actual shutdown.
func (h *SignalHandler) shutdown() {
	// Only shutdown once
	h.stateMu.Lock()
	if h.state != stateRunning {
		h.stateMu.Unlock()
		return
	}
	h.state = stateShuttingDown
	h.stateMu.Unlock()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownTimeout)
	defer cancel()

	// Stop the Fx application if set
	if h.app != nil {
		// Create a channel to receive the result of stopping the app
		done := make(chan error, 1)
		go func() {
			var err error
			switch app := h.app.(type) {
			case *fx.App:
				err = app.Stop(ctx)
			case func() error:
				err = app()
			default:
				err = fmt.Errorf("unsupported app type: %T", h.app)
			}
			done <- err
		}()

		// Wait for app to stop or timeout
		select {
		case err := <-done:
			if err != nil {
				h.logger.Error("Error stopping application", "error", err)
				h.shutdownError = err
			}
		case <-ctx.Done():
			h.logger.Error("Timeout stopping application", "error", ctx.Err())
			h.shutdownError = ctx.Err()
		}
	}

	// Close all resources with timeout
	h.resourcesMu.Lock()
	defer h.resourcesMu.Unlock()

	for _, closer := range h.resources {
		// Create a channel to receive the result of closing the resource
		done := make(chan error, 1)
		go func(c func() error) {
			done <- c()
		}(closer)

		// Wait for resource to close or timeout
		select {
		case err := <-done:
			if err != nil {
				h.logger.Error("Error closing resource", "error", err)
				h.shutdownError = err
			}
		case <-ctx.Done():
			h.logger.Error("Timeout closing resource", "error", ctx.Err())
			h.shutdownError = ctx.Err()
			return // Return from the function to stop processing more resources
		}
	}

	// Call cleanup function if set
	if h.cleanup != nil {
		// Create a channel to receive the result of cleanup
		done := make(chan struct{})
		go func() {
			h.cleanup()
			close(done)
		}()

		// Wait for cleanup to complete or timeout
		select {
		case <-done:
			// Cleanup completed successfully
		case <-ctx.Done():
			h.logger.Error("Timeout during cleanup", "error", ctx.Err())
			h.shutdownError = ctx.Err()
		}
	}

	// Update state and close done channel
	h.stateMu.Lock()
	h.state = stateShutdownComplete
	h.stateMu.Unlock()
}
