// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	signal "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

// Dependencies holds the HTTP server's dependencies
type Dependencies struct {
	fx.In

	Lifecycle    fx.Lifecycle
	Logger       common.Logger
	Config       config.Interface
	Storage      types.Interface
	IndexManager api.IndexManager
	Context      context.Context `name:"httpdContext"`
}

// Params holds the parameters required for running the HTTP server.
type Params struct {
	fx.In
	Server  *http.Server
	Logger  common.Logger
	Storage types.Interface
	Config  *config.Config
}

// serverState tracks the HTTP server's state
type serverState struct {
	mu       sync.Mutex
	started  bool
	shutdown bool
	// serverDone is closed when the server goroutine exits
	serverDone chan struct{}
}

// Server implements the HTTP server
type Server struct {
	config *config.Config
	Logger common.Logger
	server *http.Server
}

// NewServer creates a new HTTP server instance
func NewServer(params Params) *Server {
	return &Server{
		config: params.Config,
		Logger: params.Logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.Logger.Info("Starting HTTP server", "addr", s.config.Server.Address)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.Logger.Info("Stopping HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}

// Cmd represents the HTTP server command
var Cmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling with a no-op logger initially
		handler := signal.NewSignalHandler(common.NewNoOpLogger())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Track server state
		state := &serverState{
			serverDone: make(chan struct{}),
		}

		// Initialize the Fx application
		fxApp := fx.New(
			fx.NopLogger,
			common.Module,
			Module,
			fx.Provide(
				func() context.Context { return ctx },
			),
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
				// Update the signal handler with the real logger
				handler.SetLogger(p.Logger)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test storage connection
						if err := p.Storage.TestConnection(ctx); err != nil {
							return fmt.Errorf("failed to connect to storage: %w", err)
						}

						// Start HTTP server in background
						p.Logger.Info("Starting HTTP server...", "address", p.Server.Addr)
						state.mu.Lock()
						state.started = true
						state.mu.Unlock()

						go func() {
							defer close(state.serverDone)
							if err := p.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
								p.Logger.Error("HTTP server failed", "error", err)
							}
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						state.mu.Lock()
						if !state.started || state.shutdown {
							state.mu.Unlock()
							return nil
						}
						state.shutdown = true
						state.mu.Unlock()

						p.Logger.Info("Initiating graceful shutdown...")

						// Create timeout context for shutdown
						shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
						defer shutdownCancel()

						// Shutdown HTTP server
						p.Logger.Info("Shutting down HTTP server...")
						if err := p.Server.Shutdown(shutdownCtx); err != nil {
							if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
								p.Logger.Error("Error during server shutdown", "error", err)
								return fmt.Errorf("error during server shutdown: %w", err)
							}
						}

						// Wait for server goroutine to exit
						select {
						case <-state.serverDone:
							p.Logger.Info("HTTP server goroutine exited")
						case <-shutdownCtx.Done():
							p.Logger.Warn("Timeout waiting for server goroutine to exit")
						}

						p.Logger.Info("Shutdown complete")
						return nil
					},
				})
			}),
		)

		// Set the fx app for coordinated shutdown
		handler.SetFXApp(fxApp)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for shutdown signal
		handler.Wait()

		return nil
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
