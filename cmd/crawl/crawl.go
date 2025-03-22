// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// Params holds the dependencies and parameters required for the crawl operation.
type Params struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Sources    sources.Interface `name:"sourceManager"`
	Crawler    crawler.Interface
	Logger     logger.Interface
	Config     config.Interface
	Done       chan struct{}             `name:"crawlDone"`
	Context    context.Context           `name:"crawlContext"`
	Processors []models.ContentProcessor `group:"processors"`
}

// StartCrawl initializes and starts the crawling process.
func StartCrawl(p Params) error {
	source, findErr := p.Sources.FindByName(sourceName)
	if findErr != nil {
		return fmt.Errorf("error finding source: %w", findErr)
	}

	// Register lifecycle hooks
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			collectorResult, setupErr := app.SetupCollector(ctx, p.Logger, *source, p.Processors, p.Done, p.Config)
			if setupErr != nil {
				return fmt.Errorf("error setting up collector: %w", setupErr)
			}

			if configErr := app.ConfigureCrawler(p.Crawler, *source, collectorResult); configErr != nil {
				return fmt.Errorf("error configuring crawler: %w", configErr)
			}

			p.Logger.Info("Starting crawl", "source", source.Name)
			if startErr := p.Crawler.Start(ctx, source.URL); startErr != nil {
				return fmt.Errorf("error starting crawler: %w", startErr)
			}

			// Monitor crawl completion in a separate goroutine
			go func() {
				p.Crawler.Wait()
				p.Logger.Info("Crawler finished processing all URLs")

				// Signal completion through the Done channel
				close(p.Done)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping crawler")
			if stopErr := p.Crawler.Stop(ctx); stopErr != nil {
				return fmt.Errorf("error stopping crawler: %w", stopErr)
			}
			return nil
		},
	})

	return nil
}

// Cmd represents the crawl command.
var Cmd = &cobra.Command{
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

		// Create a cancellable context and set up signal handling
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		signalDone, cleanup := app.WaitForSignal(ctx, cancel)
		defer cleanup()

		// Create the crawler's completion channel
		crawlerDone := make(chan struct{})

		// Initialize the Fx application with the crawl module
		fxApp := fx.New(
			fx.NopLogger,
			Module,
			fx.Provide(
				fx.Annotate(
					func() context.Context { return ctx },
					fx.ResultTags(`name:"crawlContext"`),
				),
				fx.Annotate(
					func() chan struct{} { return crawlerDone },
					fx.ResultTags(`name:"crawlDone"`),
				),
			),
			fx.Invoke(StartCrawl),
		)

		// Start the application
		if startErr := fxApp.Start(ctx); startErr != nil {
			if errors.Is(startErr, context.Canceled) {
				return nil
			}
			return fmt.Errorf("error starting application: %w", startErr)
		}

		// Wait for completion or cancellation
		select {
		case <-signalDone:
			// Normal completion through signal
		case <-crawlerDone:
			// Crawler completed successfully
		case <-ctx.Done():
			// Context cancelled
		}

		// Perform graceful shutdown
		if shutdownErr := fxApp.Stop(ctx); shutdownErr != nil {
			return fmt.Errorf("error during shutdown: %w", shutdownErr)
		}

		return nil
	},
}

// Command returns the crawl command.
func Command() *cobra.Command {
	return Cmd
}
