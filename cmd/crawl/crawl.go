// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// crawlParams holds the dependencies for the crawl command.
type crawlParams struct {
	fx.In
	Sources     sources.Interface `name:"sourceManager"`
	Crawler     crawler.Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     storage.Interface
	Done        chan struct{}             `name:"crawlDone"`
	Context     context.Context           `name:"crawlContext"`
	Processors  []models.ContentProcessor `group:"processors"`
	SourceName  string                    `name:"sourceName"`
	ArticleChan chan *models.Article      `name:"articleChannel"`
	Handler     *signal.SignalHandler     `name:"signalHandler"`
}

// setupCrawler initializes and configures the crawler with the given source.
func setupCrawler(ctx context.Context, p crawlParams, source *sources.Config) error {
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
func startCrawler(ctx context.Context, p crawlParams, source *sources.Config, closeChannels func()) error {
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
func stopCrawler(ctx context.Context, p crawlParams) error {
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
func createCloseChannels(p crawlParams) func() {
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

			// Finally signal completion to the command
			p.Logger.Debug("Signaling completion to command", "goroutines", runtime.NumGoroutine())
			p.Handler.Complete()
			p.Logger.Debug("Signal handler completed", "goroutines", runtime.NumGoroutine())

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
		fx.Invoke(func(lc fx.Lifecycle, p crawlParams) {
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

// Command returns the crawl command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawl [source]",
		Short: "Crawl a single source defined in sources.yml",
		Long: `Crawl a single source defined in sources.yml.
The source argument must match a name defined in your sources.yml configuration file.

Example:
  gocrawl crawl example-blog`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				const errMsg = "requires exactly one source name from sources.yml\n\n" +
					"Usage:\n  %s\n\n" +
					"Run 'gocrawl list' to see available sources"
				return fmt.Errorf(errMsg, cmd.UseLine())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName = args[0]
			log, err := logger.NewDevelopmentLogger("debug")
			if err != nil {
				return fmt.Errorf("error creating logger: %w", err)
			}
			log.Info("Starting crawl command", "source", sourceName, "goroutines", runtime.NumGoroutine())

			// Use Cobra's context for proper command lifecycle management
			cmdCtx := cmd.Context()
			handler := signal.NewSignalHandler(log)
			cleanup := handler.Setup(cmdCtx)
			defer cleanup()

			fxApp := createFXApp(cmdCtx, sourceName, handler)
			handler.SetFXApp(fxApp)

			// Start the application
			log.Debug("Starting fx application", "goroutines", runtime.NumGoroutine())

			if errStart := fxApp.Start(cmdCtx); errStart != nil {
				if errors.Is(errStart, context.Canceled) {
					log.Debug("Application start cancelled", "goroutines", runtime.NumGoroutine())
					return nil
				}
				return fmt.Errorf("error starting application: %w", errStart)
			}

			log.Debug("Fx application started successfully", "goroutines", runtime.NumGoroutine())

			// Wait for completion or signal
			log.Debug(
				"Waiting for completion or signal",
				"handler_state", handler.GetState(),
				"goroutines", runtime.NumGoroutine(),
			)

			if !handler.Wait() {
				log.Debug("Context cancelled, initiating shutdown", "goroutines", runtime.NumGoroutine())
				// Stop the application since context was cancelled
				if errStop := fxApp.Stop(cmdCtx); errStop != nil {
					log.Error("Error stopping application", "error", errStop, "goroutines", runtime.NumGoroutine())
					return fmt.Errorf("error stopping application: %w", errStop)
				}
				log.Debug("Fx application stopped successfully", "goroutines", runtime.NumGoroutine())
				return nil
			}

			log.Debug(
				"Signal handler wait completed",
				"handler_state", handler.GetState(),
				"goroutines", runtime.NumGoroutine(),
			)

			// Exit the command
			log.Debug(
				"Command execution complete, exiting",
				"handler_state", handler.GetState(),
				"goroutines", runtime.NumGoroutine(),
			)
			return nil
		},
	}

	return cmd
}
