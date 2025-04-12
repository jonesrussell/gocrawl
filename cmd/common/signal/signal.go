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
	for i, closer := range rm.resources {
		done := make(chan error, 1)
		go func(idx int, c func() error) {
			err := c()
			if err == nil {
				logger.Info("Resource closed successfully", "index", idx)
			} else {
				logger.Error("Failed to close resource", "index", idx, "error", err)
			}
			done <- err
		}(i, closer)

		select {
		case err := <-done:
			if err != nil {
				lastErr = err
			}
		case <-ctx.Done():
			logger.Error("Timeout closing resource", "index", i, "error", ctx.Err())
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

// SignalHandler handles OS signals and application shutdown.
type SignalHandler struct {
	logger          logger.Interface
	Done            chan struct{}
	shutdownChan    chan struct{}
	state           shutdownState
	stateMu         sync.RWMutex
	shutdownTimeout time.Duration
	shutdownError   error
	resourceManager *ResourceManager
	appManager      *AppManager
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewSignalHandler creates a new signal handler.
func NewSignalHandler(logger logger.Interface) *SignalHandler {
	return &SignalHandler{
		logger:          logger,
		Done:            make(chan struct{}),
		shutdownChan:    make(chan struct{}),
		state:           stateRunning,
		shutdownTimeout: DefaultShutdownTimeout,
		resourceManager: NewResourceManager(),
		appManager:      NewAppManager(),
	}
}

// Setup sets up signal handling and returns a cleanup function.
func (h *SignalHandler) Setup(ctx context.Context) func() {
	h.ctx, h.cancel = context.WithCancel(ctx)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handling goroutine
	go func() {
		defer signal.Stop(sigChan)
		defer close(sigChan)

		for sig := range sigChan {
			h.logger.Info("Received signal", "signal", sig)
			if !h.IsShuttingDown() {
				h.RequestShutdown()
				return
			} else {
				h.logger.Info("Already shutting down, ignoring signal", "signal", sig)
			}
		}
	}()

	return func() {
		signal.Stop(sigChan)
		close(sigChan)
	}
}

// RequestShutdown initiates a graceful shutdown.
func (h *SignalHandler) RequestShutdown() {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()

	if h.state == stateRunning {
		h.state = stateShuttingDown
		h.logger.Info("Initiating graceful shutdown")
		// Cancel the context first
		if h.cancel != nil {
			h.cancel()
		}
		// Signal shutdown
		close(h.shutdownChan)
		// Start shutdown process
		go h.shutdown()
	}
}

// Wait waits for shutdown to complete.
func (h *SignalHandler) Wait() error {
	<-h.Done
	return h.shutdownError
}

// GetState returns the current state of the handler.
func (h *SignalHandler) GetState() string {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()

	switch h.state {
	case stateRunning:
		return "running"
	case stateShuttingDown:
		return "shutting down"
	case stateShutdownComplete:
		return "shutdown complete"
	default:
		return "unknown"
	}
}

// SetLogger sets the logger for the handler.
func (h *SignalHandler) SetLogger(logger logger.Interface) {
	h.logger = logger
}

// IsShuttingDown returns whether the handler is shutting down.
func (h *SignalHandler) IsShuttingDown() bool {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.state == stateShuttingDown
}

// SetShutdownTimeout sets the shutdown timeout.
func (h *SignalHandler) SetShutdownTimeout(timeout time.Duration) {
	h.shutdownTimeout = timeout
}

// AddResource adds a resource to be closed during shutdown.
func (h *SignalHandler) AddResource(closer func() error) {
	h.resourceManager.AddResource(closer)
}

// SetCleanup sets the cleanup function.
func (h *SignalHandler) SetCleanup(cleanup func()) {
	h.resourceManager.SetCleanup(cleanup)
}

// SetFXApp sets the Fx application for shutdown.
func (h *SignalHandler) SetFXApp(app any) {
	h.appManager.SetApp(app)
}

// shutdown performs the actual shutdown process.
func (h *SignalHandler) shutdown() {
	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(h.ctx, h.shutdownTimeout)
	defer cancel()

	// Wait for shutdown signal or timeout
	select {
	case <-h.shutdownChan:
		// Continue with shutdown
	case <-ctx.Done():
		h.logger.Error("Shutdown timed out", "error", ctx.Err())
		h.shutdownError = ctx.Err()
		h.stateMu.Lock()
		h.state = stateShutdownComplete
		h.stateMu.Unlock()
		close(h.Done)
		return
	}

	// Close resources
	if err := h.resourceManager.CloseResources(ctx, h.logger); err != nil {
		h.logger.Error("Error closing resources", "error", err)
		h.shutdownError = err
	}

	// Stop the Fx application if set
	if err := h.appManager.StopApp(ctx, h.logger); err != nil {
		h.logger.Error("Error stopping application", "error", err)
		h.shutdownError = err
	}

	// Mark shutdown as complete
	h.stateMu.Lock()
	h.state = stateShutdownComplete
	h.stateMu.Unlock()

	// Signal completion
	close(h.Done)
}

// StateChan returns a channel that emits the current state of the handler.
// The channel will be closed when the handler is shutdown.
func (h *SignalHandler) StateChan() <-chan string {
	stateChan := make(chan string, 1)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		defer close(stateChan)

		for {
			select {
			case <-ticker.C:
				state := h.GetState()
				select {
				case stateChan <- state:
					if state == "shutdown complete" {
						return
					}
				default:
					// Skip if channel is full
				}
			case <-h.Done:
				return
			}
		}
	}()

	return stateChan
}
