// Package signal provides signal handling functionality.
package signal

import (
	"context"
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
	SetFXApp(app *fx.App)
	RequestShutdown()
	Wait() bool
	AddResource(closer func() error)
	GetState() string
	SetLogger(logger logger.Interface)
	SetCleanup(cleanup func())
	IsShuttingDown() bool
}

// SignalHandler handles OS signals and graceful shutdown.
type SignalHandler struct {
	logger          logger.Interface
	app             *fx.App
	done            chan struct{}
	testMode        bool
	state           shutdownState
	stateMu         sync.RWMutex
	resources       []func() error
	resourcesMu     sync.Mutex
	shutdownTimeout time.Duration
	cleanup         func()
	shutdownError   error
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
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Start signal handling goroutine
	go func() {
		select {
		case <-ctx.Done():
			h.logger.Info("Context cancelled, initiating shutdown")
			h.shutdown()
		case sig := <-sigChan:
			h.logger.Info("Received signal, initiating shutdown", "signal", sig)
			h.shutdown()
		}
	}()

	// Return cleanup function
	return func() {
		signal.Stop(sigChan)
		close(sigChan)
	}
}

// SetFXApp sets the Fx application for coordinated shutdown.
func (h *SignalHandler) SetFXApp(app *fx.App) {
	h.app = app
}

// RequestShutdown requests a graceful shutdown.
func (h *SignalHandler) RequestShutdown() {
	h.shutdown()
}

// Wait waits for shutdown to complete and returns true if shutdown completed successfully.
func (h *SignalHandler) Wait() bool {
	<-h.done
	return h.shutdownError == nil
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

// SetTestMode enables test mode.
func (h *SignalHandler) SetTestMode(enabled bool) {
	h.testMode = enabled
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
		if err := h.app.Stop(ctx); err != nil {
			h.logger.Error("Error stopping application", "error", err)
			h.shutdownError = err
		}
	}

	// Close all resources
	h.resourcesMu.Lock()
	for _, closer := range h.resources {
		if err := closer(); err != nil {
			h.logger.Error("Error closing resource", "error", err)
			h.shutdownError = err
		}
	}
	h.resourcesMu.Unlock()

	// Call cleanup function if set
	if h.cleanup != nil {
		h.cleanup()
	}

	// Update state and close done channel
	h.stateMu.Lock()
	h.state = stateShutdownComplete
	h.stateMu.Unlock()
	close(h.done)
}
