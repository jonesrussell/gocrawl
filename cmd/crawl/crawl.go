// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// SetSourceName sets the source name for testing purposes
func SetSourceName(name string) {
	sourceName = name
}

// Params holds the dependencies and parameters required for the crawl operation.
type Params struct {
	fx.In

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// Sources provides access to configured crawl sources
	Sources *sources.Sources

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Processors is a slice of content processors, injected as a group
	Processors []models.ContentProcessor `group:"processors"`

	// Logger provides structured logging capabilities
	Logger logger.Interface

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`

	// Config holds the application configuration
	Config config.Interface

	// Context provides the context for the crawl operation
	Context context.Context `name:"crawlContext"`
}

// StartCrawl initializes and starts the crawling process.
func StartCrawl(p Params) error {
	return startCrawl(p)
}

// startCrawl is the internal implementation of StartCrawl
func startCrawl(p Params) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Get the source configuration first
	source, err := p.Sources.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("error finding source: %w", err)
	}

	// Set up the collector
	collectorResult, err := app.SetupCollector(p.Context, p.Logger, *source, p.Processors, p.Done, p.Config)
	if err != nil {
		return fmt.Errorf("error setting up collector: %w", err)
	}

	// Configure the crawler
	if err := app.ConfigureCrawler(p.CrawlerInstance, *source, collectorResult); err != nil {
		return fmt.Errorf("error configuring crawler: %w", err)
	}

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

		// Create a parent context that can be cancelled
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling
		done, cleanup := app.WaitForSignal(ctx, cancel)
		defer cleanup()

		// Initialize the Fx application with required modules and dependencies
		fxApp := fx.New(
			common.Module,
			article.Module,
			content.Module,
			collector.Module(),
			Module,
			fx.Provide(
				fx.Annotate(
					func() chan struct{} {
						return make(chan struct{})
					},
					fx.ResultTags(`name:"crawlDone"`),
				),
				fx.Annotate(
					func() context.Context {
						return ctx
					},
					fx.ResultTags(`name:"crawlContext"`),
				),
				fx.Annotate(
					func() string {
						return sourceName
					},
					fx.ResultTags(`name:"sourceName"`),
				),
				fx.Annotate(
					func(sources *sources.Sources) (string, string) {
						source, err := sources.FindByName(sourceName)
						if err != nil {
							return "", ""
						}
						return source.Index, source.ArticleIndex
					},
					fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
				),
				func() chan *models.Article {
					return make(chan *models.Article, app.DefaultChannelBufferSize)
				},
			),
		)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for completion or interruption
		<-done

		// Perform graceful shutdown
		return app.GracefulShutdown(fxApp)
	},
}

// Command returns the crawl command.
func Command() *cobra.Command {
	return Cmd
}
