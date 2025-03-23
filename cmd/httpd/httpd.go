// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout = 30 * time.Second
	// componentShutdownTimeout is the maximum time to wait for individual components
	componentShutdownTimeout = 2 * time.Second
	// expectedGoroutines is the expected number of goroutines after cleanup
	expectedGoroutines = 3 // main + signal handler + fx signal receiver
)

// Params holds the parameters required for running the HTTP server.
type Params struct {
	fx.In
	Server  *http.Server
	Logger  logger.Interface
	Storage storage.Interface
}

// shutdownContext wraps a context with additional shutdown information
type shutdownContext struct {
	context.Context
	log logger.Interface
}

func newShutdownContext(ctx context.Context, log logger.Interface) *shutdownContext {
	return &shutdownContext{
		Context: ctx,
		log:     log,
	}
}

func (sc *shutdownContext) logShutdownProgress(phase string) {
	sc.log.Debug("Shutdown progress", "phase", phase)
}

func cleanupGoroutines(log logger.Interface) {
	count := runtime.NumGoroutine()
	if count > expectedGoroutines {
		log.Warn("Some goroutines may not have cleaned up properly",
			"count", count,
			"expected", expectedGoroutines,
		)
	} else {
		log.Debug("All goroutines cleaned up",
			"count", count,
			"expected", expectedGoroutines,
		)
	}
}

// Cmd represents the HTTP server command
var Cmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		errChan := make(chan error, 1)
		doneChan := make(chan struct{})
		var log logger.Interface

		app := fx.New(
			fx.Provide(
				func() context.Context { return ctx },
			),
			common.Module,
			Module,
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
				log = p.Logger
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						if err := p.Storage.TestConnection(ctx); err != nil {
							return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
						}

						p.Logger.Info("Starting HTTP server...", "address", p.Server.Addr)
						go func() {
							defer close(errChan)
							defer close(doneChan)

							if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								p.Logger.Error("HTTP server failed", "error", err)
								errChan <- err
								return
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						sc := newShutdownContext(ctx, p.Logger)
						sc.logShutdownProgress("start")

						shutdownCtx, cancel := context.WithTimeout(sc, componentShutdownTimeout)
						defer cancel()

						sc.logShutdownProgress("http_server")
						if err := p.Server.Shutdown(shutdownCtx); err != nil {
							p.Logger.Error("Error during server shutdown", "error", err)
							return err
						}

						sc.logShutdownProgress("storage")
						if err := p.Storage.Close(); err != nil {
							p.Logger.Error("Error closing storage connection", "error", err)
						}

						sc.logShutdownProgress("complete")
						return nil
					},
				})
			}),
		)

		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		var serverErr error
		select {
		case serverErr = <-errChan:
			log.Debug("Server error received")
			stopCtx, cancel := context.WithTimeout(cmd.Context(), shutdownTimeout)
			defer cancel()
			if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("Error stopping application", "error", err)
			}
			<-app.Done()
		case <-doneChan:
			log.Info("Server stopped normally")
			stopCtx, cancel := context.WithTimeout(cmd.Context(), shutdownTimeout)
			defer cancel()
			if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("Error stopping application", "error", err)
			}
			<-app.Done()
		case <-ctx.Done():
			sc := newShutdownContext(cmd.Context(), log)
			sc.logShutdownProgress("signal_received")

			shutdownCtx, cancel := context.WithTimeout(sc, shutdownTimeout)
			defer cancel()

			sc.logShutdownProgress("app_stop")
			if err := app.Stop(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("Error stopping application", "error", err)
			}

			select {
			case <-app.Done():
				sc.logShutdownProgress("normal_complete")
			case <-time.After(componentShutdownTimeout):
				sc.logShutdownProgress("timeout")
				stop()
			}

			cleanupGoroutines(log)
			log.Info("Application shutdown complete")
		}

		stop()

		if serverErr != nil && !errors.Is(serverErr, context.Canceled) && !errors.Is(serverErr, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", serverErr)
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}

		return nil
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
