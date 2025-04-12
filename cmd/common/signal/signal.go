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

const DefaultShutdownTimeout = 30 * time.Second

type shutdownState int

const (
	stateRunning shutdownState = iota
	stateShuttingDown
	stateShutdownComplete
)

type Interface interface {
	Setup(ctx context.Context) func()
	SetFXApp(app any)
	RequestShutdown()
	Wait() error
	AddResource(closer func() error)
	GetState() string
	SetLogger(logger logger.Interface)
	SetCleanup(cleanup func())
	IsShuttingDown() bool
}

type SignalHandler struct {
	logger          logger.Interface
	done            chan struct{}
	mu              sync.Mutex
	app             any
	state           shutdownState
	stateMu         sync.RWMutex
	resources       []func() error
	resourcesMu     sync.Mutex
	shutdownTimeout time.Duration
	cleanup         func()
	shutdownError   error
}

// NewSignalHandler creates a new SignalHandler instance.
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		logger:          logger,
		done:            make(chan struct{}),
		state:           stateRunning,
		shutdownTimeout: DefaultShutdownTimeout,
	}
}

// Setup initializes signal handling.
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

// Wait blocks until shutdown completes.
func (h *SignalHandler) Wait() error {
	<-h.done
	return h.shutdownError
}

// RequestShutdown safely initiates shutdown.
func (h *SignalHandler) RequestShutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state != stateRunning {
		return
	}
	h.shutdown()
	close(h.done)
}

// shutdown performs the actual shutdown.
func (h *SignalHandler) shutdown() {
	h.transitionState(stateShuttingDown)

	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownTimeout)
	defer cancel()

	h.stopApp(ctx)
	h.closeResources(ctx)

	if h.cleanup != nil {
		h.logger.Info("Performing cleanup")
		h.cleanup()
	}

	h.transitionState(stateShutdownComplete)
}

// stopApp stops the fx app or any other application.
func (h *SignalHandler) stopApp(ctx context.Context) {
	if h.app == nil {
		return
	}

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

// closeResources safely closes all resources.
func (h *SignalHandler) closeResources(ctx context.Context) {
	h.resourcesMu.Lock()
	defer h.resourcesMu.Unlock()

	for _, closer := range h.resources {
		done := make(chan error, 1)
		go func(c func() error) { done <- c() }(closer)

		select {
		case err := <-done:
			if err != nil {
				h.logger.Error("Error closing resource", "error", err)
				h.shutdownError = err
			}
		case <-ctx.Done():
			h.logger.Error("Timeout closing resource", "error", ctx.Err())
			h.shutdownError = ctx.Err()
			return
		}
	}
}

// transitionState safely updates the shutdown state.
func (h *SignalHandler) transitionState(newState shutdownState) {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()
	h.state = newState
}

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

// SetFXApp sets the Fx app for coordinated shutdown.
func (h *SignalHandler) SetFXApp(app any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.app = app
}

// AddResource registers a resource for graceful shutdown.
func (h *SignalHandler) AddResource(closer func() error) {
	h.resourcesMu.Lock()
	defer h.resourcesMu.Unlock()
	h.resources = append(h.resources, closer)
}

// SetLogger updates the logger.
func (h *SignalHandler) SetLogger(logger logger.Interface) {
	h.logger = logger
}

// SetCleanup registers a cleanup function.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.cleanup = cleanup
}

// IsShuttingDown checks if shutdown is in progress.
func (h *SignalHandler) IsShuttingDown() bool {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.state == stateShuttingDown
}

// SetShutdownTimeout updates the shutdown timeout duration.
func (h *SignalHandler) SetShutdownTimeout(timeout time.Duration) {
	h.shutdownTimeout = timeout
}
