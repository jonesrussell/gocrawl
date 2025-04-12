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

// Interface defines the contract for signal handling.
type Interface interface {
	Setup(ctx context.Context) func()
	RequestShutdown()
	Wait() error
	GetState() string
	SetLogger(logger logger.Interface)
	IsShuttingDown() bool
}

// ResourceManager handles resource cleanup during shutdown.
type ResourceManager struct {
	mu        sync.Mutex
	resources []func() error
	cleanup   func()
}

// NewResourceManager creates a new ResourceManager instance.
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make([]func() error, 0),
	}
}

// AddResource registers a resource for graceful shutdown.
func (rm *ResourceManager) AddResource(closer func() error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.resources = append(rm.resources, closer)
}

// SetCleanup registers a cleanup function.
func (rm *ResourceManager) SetCleanup(cleanup func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.cleanup = cleanup
}

// CloseResources safely closes all resources.
func (rm *ResourceManager) CloseResources(ctx context.Context, logger logger.Interface) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	var lastErr error
	for _, closer := range rm.resources {
		done := make(chan error, 1)
		go func(c func() error) { done <- c() }(closer)

		select {
		case err := <-done:
			if err != nil {
				logger.Error("Error closing resource", "error", err)
				lastErr = err
			}
		case <-ctx.Done():
			logger.Error("Timeout closing resource", "error", ctx.Err())
			return ctx.Err()
		}
	}

	if rm.cleanup != nil {
		logger.Info("Performing cleanup")
		rm.cleanup()
	}

	return lastErr
}

// AppManager handles application lifecycle during shutdown.
type AppManager struct {
	mu  sync.Mutex
	app any
}

// NewAppManager creates a new AppManager instance.
func NewAppManager() *AppManager {
	return &AppManager{}
}

// SetApp sets the application for coordinated shutdown.
func (am *AppManager) SetApp(app any) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.app = app
}

// StopApp stops the application gracefully.
func (am *AppManager) StopApp(ctx context.Context, logger logger.Interface) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.app == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() {
		var err error
		switch app := am.app.(type) {
		case *fx.App:
			err = app.Stop(ctx)
		case func() error:
			err = app()
		default:
			err = fmt.Errorf("unsupported app type: %T", am.app)
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Error("Error stopping application", "error", err)
			return err
		}
	case <-ctx.Done():
		logger.Error("Timeout stopping application", "error", ctx.Err())
		return ctx.Err()
	}

	return nil
}

// SignalHandler implements the signal handling interface.
type SignalHandler struct {
	logger          logger.Interface
	done            chan struct{}
	mu              sync.Mutex
	state           shutdownState
	stateMu         sync.RWMutex
	shutdownTimeout time.Duration
	shutdownError   error
	resourceManager *ResourceManager
	appManager      *AppManager
}

// NewSignalHandler creates a new SignalHandler instance.
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		logger:          logger,
		done:            make(chan struct{}),
		state:           stateRunning,
		shutdownTimeout: DefaultShutdownTimeout,
		resourceManager: NewResourceManager(),
		appManager:      NewAppManager(),
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

	// Stop the application first
	if err := h.appManager.StopApp(ctx, h.logger); err != nil {
		h.shutdownError = err
	}

	// Then close resources
	if err := h.resourceManager.CloseResources(ctx, h.logger); err != nil {
		h.shutdownError = err
	}

	h.transitionState(stateShutdownComplete)
}

// transitionState safely updates the shutdown state.
func (h *SignalHandler) transitionState(newState shutdownState) {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()
	h.state = newState
}

// GetState returns the current shutdown state as a string.
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

// SetLogger updates the logger.
func (h *SignalHandler) SetLogger(logger logger.Interface) {
	h.logger = logger
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

// AddResource registers a resource for graceful shutdown.
func (h *SignalHandler) AddResource(closer func() error) {
	h.resourceManager.AddResource(closer)
}

// SetCleanup registers a cleanup function.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.resourceManager.SetCleanup(cleanup)
}

// SetFXApp sets the Fx app for coordinated shutdown.
func (h *SignalHandler) SetFXApp(app any) {
	h.appManager.SetApp(app)
}
