// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	// #nosec G108 -- Profiling endpoint is intentionally exposed for debugging
	_ "net/http/pprof"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

const (
	// pprofServerReadTimeout is the maximum duration for reading the entire request.
	pprofServerReadTimeout = 10 * time.Second
	// pprofServerWriteTimeout is the maximum duration for writing the entire response.
	pprofServerWriteTimeout = 10 * time.Second
	// pprofServerIdleTimeout is the maximum duration for keeping idle connections alive.
	pprofServerIdleTimeout = 120 * time.Second
	// stackTraceBufferSize is the size of the buffer used for stack traces (64KB).
	stackTraceBufferSize = 1 << 16
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// Dependencies holds the crawl command's dependencies
type Dependencies struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Sources     sources.Interface `name:"sourceManager"`
	Crawler     crawler.Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     types.Interface
	Done        chan struct{}             `name:"crawlDone"`
	Context     context.Context           `name:"crawlContext"`
	Processors  []models.ContentProcessor `group:"processors"`
	SourceName  string                    `name:"sourceName"`
	ArticleChan chan *models.Article      `name:"articleChannel"`
}

// setupCrawler initializes and configures the crawler with the given source.
func setupCrawler(ctx context.Context, p Dependencies, source *sources.Config) error {
	p.Logger.Debug("Setting up collector", "source", source.Name)
	collectorResult, setupErr := app.SetupCollector(ctx, p.Logger, *source, p.Processors, p.Done, p.Config)
	if setupErr != nil {
		return fmt.Errorf("error setting up collector: %w", setupErr)
	}

	p.Logger.Debug("Configuring crawler", "source", source.Name)
	if configErr := app.ConfigureCrawler(p.Crawler, *source, collectorResult); configErr != nil {
		return fmt.Errorf("error configuring crawler: %w", configErr)
	}

	return nil
}

// startCrawler starts the crawler and monitors its completion.
func startCrawler(ctx context.Context, p Dependencies, source *sources.Config, closeChannels func()) error {
	p.Logger.Info("Starting crawl", "source", source.Name, "url", source.URL)
	if startErr := p.Crawler.Start(ctx, source.URL); startErr != nil {
		return fmt.Errorf("error starting crawler: %w", startErr)
	}

	p.Logger.Debug("Waiting for crawler to finish", "source", source.Name)
	p.Crawler.Wait()
	p.Logger.Info("Crawler finished processing all URLs", "source", source.Name)
	p.Logger.Debug("Initiating channel closure sequence")
	closeChannels()
	p.Logger.Debug("Channel closure sequence initiated")

	return nil
}

// stopCrawler gracefully stops the crawler.
func stopCrawler(ctx context.Context, p Dependencies) error {
	p.Logger.Info("Stopping crawler", "source", p.SourceName)
	if stopErr := p.Crawler.Stop(ctx); stopErr != nil {
		if errors.Is(stopErr, context.Canceled) {
			p.Logger.Info("Crawler stop was cancelled", "source", p.SourceName)
			return nil
		}
		return fmt.Errorf("error stopping crawler: %w", stopErr)
	}
	p.Logger.Debug("Crawler stopped successfully", "source", p.SourceName)
	return nil
}

// createCloseChannels creates a function to safely close all channels.
func createCloseChannels(p Dependencies) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			p.Logger.Debug("Starting channel closure sequence", "goroutines", runtime.NumGoroutine())
			// First close the article channel to signal no more articles
			p.Logger.Debug("Closing article channel", "goroutines", runtime.NumGoroutine())
			close(p.ArticleChan)
			p.Logger.Debug("Article channel closed", "goroutines", runtime.NumGoroutine())

			// Then close the Done channel to trigger OnStop hooks
			p.Logger.Debug("Closing Done channel", "goroutines", runtime.NumGoroutine())
			close(p.Done)
			p.Logger.Debug("Done channel closed", "goroutines", runtime.NumGoroutine())

			p.Logger.Debug("Shutdown sequence completed", "source", p.SourceName, "goroutines", runtime.NumGoroutine())
		})
	}
}

// createFXApp creates and configures the fx application.
func createFXApp(ctx context.Context, sourceName string, handler *signal.SignalHandler) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger {
			log, _ := zap.NewDevelopment()
			return &fxevent.ZapLogger{Logger: log}
		}),
		Module,
		fx.Provide(
			fx.Annotate(
				func() context.Context { return ctx },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return sourceName },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return handler },
				fx.ResultTags(`name:"signalHandler"`),
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, p Dependencies) {
			closeChannels := createCloseChannels(p)

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					source, findErr := p.Sources.FindByName(p.SourceName)
					if findErr != nil {
						return fmt.Errorf("error finding source: %w", findErr)
					}

					if err := setupCrawler(ctx, p, source); err != nil {
						return err
					}

					// Start the crawler in a goroutine managed by Fx
					go func() {
						if err := startCrawler(ctx, p, source, closeChannels); err != nil {
							p.Logger.Error("Error in crawler goroutine", "error", err)
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					if err := stopCrawler(ctx, p); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)
}

// formatGoroutineTrace formats a single goroutine trace for logging.
func formatGoroutineTrace(trace string) (int, string, string) {
	lines := strings.Split(trace, "\n")
	if len(lines) == 0 {
		return 0, "unknown", trace
	}

	// Extract goroutine ID from the first line
	id := 0
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) > 1 {
			// Extract number from "goroutine N"
			_, err := fmt.Sscanf(parts[1], "%d", &id)
			if err != nil {
				return 0, "", ""
			}
		}
	}

	// Extract state from the second line
	state := "unknown"
	if len(lines) > 1 {
		parts := strings.Split(lines[1], ":")
		if len(parts) > 1 {
			state = strings.TrimSpace(parts[1])
		}
	}

	return id, state, trace
}

// printGoroutines prints the number of active goroutines and their stack traces.
func printGoroutines(log common.Logger) {
	count := runtime.NumGoroutine()
	log.Debug("Active goroutines", "count", count)

	buf := make([]byte, stackTraceBufferSize)
	n := runtime.Stack(buf, true)
	if n == 0 {
		return
	}

	traces := string(buf[:n])
	goroutines := strings.Split(traces, "\n\n")

	for _, trace := range goroutines {
		if trace == "" {
			continue
		}
		id, state, fullTrace := formatGoroutineTrace(trace)
		log.Debug("Goroutine",
			"id", id,
			"state", state,
			"trace", fullTrace,
		)
	}
}

// setupPprofServer creates and starts a pprof server.
func setupPprofServer(log common.Logger) (*http.Server, func()) {
	server := &http.Server{
		Addr:         "localhost:6060",
		ReadTimeout:  pprofServerReadTimeout,
		WriteTimeout: pprofServerWriteTimeout,
		IdleTimeout:  pprofServerIdleTimeout,
	}

	// Create a channel to signal server shutdown
	done := make(chan struct{})

	// Start server in a goroutine
	go func() {
		log.Info("Starting pprof server on localhost:6060")
		if serverErr := server.ListenAndServe(); serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			log.Error("pprof server error", "error", serverErr)
		}
		close(done)
	}()

	// Return cleanup function
	cleanup := func() {
		log.Debug("Shutting down pprof server")
		if err := server.Close(); err != nil {
			log.Error("Error closing pprof server", "error", err)
		}
		<-done // Wait for server to finish
		log.Debug("Pprof server shutdown complete")
	}

	return server, cleanup
}

// handleContextCancellation handles the case when the context is cancelled.
func handleContextCancellation(ctx context.Context, log common.Logger, fxApp *fx.App) error {
	log.Debug("Context cancelled, initiating shutdown", "goroutines", runtime.NumGoroutine())
	printGoroutines(log)
	if errStop := fxApp.Stop(ctx); errStop != nil {
		log.Error("Error stopping application", "error", errStop, "goroutines", runtime.NumGoroutine())
		return fmt.Errorf("error stopping application: %w", errStop)
	}
	log.Debug("Fx application stopped successfully", "goroutines", runtime.NumGoroutine())
	printGoroutines(log)
	return nil
}

// Cmd represents the crawl command
var Cmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
You can specify the source to crawl using the --source flag.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling with a no-op logger initially
		handler := signal.NewSignalHandler(logger.NewNoOp())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Initialize the Fx application
		fxApp := fx.New(
			fx.NopLogger,
			common.Module,
			crawler.Module,
			fx.Provide(
				func() context.Context { return ctx },
			),
			fx.Invoke(func(lc fx.Lifecycle, p Dependencies) {
				// Update the signal handler with the real logger
				handler.SetLogger(p.Logger)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test storage connection
						if err := p.Storage.TestConnection(ctx); err != nil {
							return fmt.Errorf("failed to connect to storage: %w", err)
						}

						// Start crawler
						p.Logger.Info("Starting crawler...", "source", p.SourceName)
						if err := p.Crawler.Start(ctx, p.SourceName); err != nil {
							return fmt.Errorf("failed to start crawler: %w", err)
						}

						return nil
					},
					OnStop: func(ctx context.Context) error {
						p.Logger.Info("Stopping crawler...")
						if err := p.Crawler.Stop(ctx); err != nil {
							return fmt.Errorf("failed to stop crawler: %w", err)
						}

						// Close storage connection
						p.Logger.Info("Closing storage connection...")
						if err := p.Storage.Close(); err != nil {
							p.Logger.Error("Error closing storage connection", "error", err)
							// Don't return error here as crawler is already stopped
						}

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

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
