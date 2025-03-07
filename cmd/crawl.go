package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// DefaultChannelBufferSize is the default size for buffered channels used for
	// processing articles and content during crawling.
	DefaultChannelBufferSize = 100
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// CrawlParams holds the dependencies and parameters required for the crawl operation.
// It uses fx.In to indicate that these fields should be injected by the fx dependency
// injection framework.
type CrawlParams struct {
	fx.In

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// Sources provides access to configured crawl sources
	Sources *sources.Sources

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Processors is a slice of content processors, injected as a group
	// The first processor handles articles, the second handles content
	Processors []models.ContentProcessor `group:"processors"`

	// Logger provides structured logging capabilities
	Logger common.Logger

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`
}

// crawlCmd represents the crawl command that initiates the crawling process for a
// specified source. It reads the source configuration from sources.yml and starts
// the crawling process with the configured parameters.
var crawlCmd = &cobra.Command{
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
		// Create a cancellable context for managing the application lifecycle
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sourceName = args[0]

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create a channel to receive the done signal when crawling completes
		doneChan := make(chan struct{})

		// Create an Fx application with all required dependencies
		app := fx.New(
			// Supply the done channel to the application
			fx.Supply(fx.Annotate(
				doneChan,
				fx.ResultTags(`name:"crawlDone"`),
			)),
			fx.Provide(
				// Provide buffered channels for article and content processing
				func() chan *models.Article {
					return make(chan *models.Article, DefaultChannelBufferSize)
				},
				func() chan *models.Content {
					return make(chan *models.Content, DefaultChannelBufferSize)
				},
				// Provide the source name from command line argument
				fx.Annotate(
					func() string {
						return sourceName
					},
					fx.ResultTags(`name:"sourceName"`),
				),
				// Provide the article index name from source configuration
				fx.Annotate(
					func(s *sources.Sources) string {
						source, _ := s.FindByName(sourceName)
						if source != nil {
							return source.ArticleIndex
						}
						return ""
					},
					fx.ResultTags(`name:"indexName"`),
				),
				// Provide the content index name from source configuration
				fx.Annotate(
					func(s *sources.Sources) string {
						source, _ := s.FindByName(sourceName)
						if source != nil {
							return source.Index
						}
						return ""
					},
					fx.ResultTags(`name:"contentIndex"`),
				),
				// Provide the article processor
				fx.Annotate(
					func(
						logger common.Logger,
						service article.Interface,
						storage common.Storage,
						params struct {
							fx.In
							IndexName string `name:"indexName"`
						},
					) models.ContentProcessor {
						return &article.Processor{
							Logger:         logger,
							ArticleService: service,
							Storage:        storage,
							IndexName:      params.IndexName,
						}
					},
					fx.ResultTags(`group:"processors"`),
				),
				// Provide the content processor
				fx.Annotate(
					func(
						service content.Interface,
						storage common.Storage,
						logger common.Logger,
						params struct {
							fx.In
							IndexName string `name:"contentIndex"`
						},
					) models.ContentProcessor {
						return content.NewProcessor(service, storage, logger, params.IndexName)
					},
					fx.ResultTags(`group:"processors"`),
				),
			),
			// Include required modules
			common.Module,
			crawler.Module,
			article.Module,
			content.Module,
			fx.Invoke(startCrawl),
		)

		// Start the application and handle any startup errors
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either:
		// - A signal interrupt (SIGINT/SIGTERM)
		// - Context cancellation
		// - Crawl completion (doneChan closed)
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
		case <-ctx.Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		case <-doneChan:
			common.PrintInfof("\nCrawl completed successfully, initiating shutdown...")
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultOperationTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
			return err
		}

		return nil
	},
}

// startCrawl initializes and starts the crawling process. It sets up the crawler
// with the specified source configuration and manages the crawl lifecycle.
//
// Parameters:
//   - p: CrawlParams containing all required dependencies and configuration
//
// Returns:
//   - error: Any error that occurred during setup or initialization
func startCrawl(p CrawlParams) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Get the crawler instance to access index service
	crawler, ok := p.CrawlerInstance.(*crawler.Crawler)
	if !ok {
		return errors.New("crawler instance is not of type *crawler.Crawler")
	}

	// Configure the sources with crawler and index manager
	p.Sources.SetCrawler(p.CrawlerInstance)
	p.Sources.SetIndexManager(crawler.IndexService)

	// Get the source configuration
	source, err := p.Sources.FindByName(sourceName)
	if err != nil {
		return err
	}

	// Parse and validate rate limit from configuration
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	// Create and configure the collector
	collectorResult, err := collector.New(collector.Params{
		BaseURL:          source.URL,
		MaxDepth:         source.MaxDepth,
		RateLimit:        rateLimit,
		Debugger:         logger.NewCollyDebugger(p.Logger),
		Logger:           p.Logger,
		ArticleProcessor: p.Processors[0], // First processor handles articles
		ContentProcessor: p.Processors[1], // Second processor handles content
		Source:           source,
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	crawler.SetCollector(collectorResult.Collector)

	// Configure lifecycle hooks for crawl management
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the crawl in a goroutine to not block
			go func() {
				if startErr := p.Sources.Start(ctx, sourceName); startErr != nil {
					if !errors.Is(startErr, context.Canceled) {
						p.Logger.Error("Crawl failed", "error", startErr)
					}
				}
				// Signal completion by closing the done channel
				close(p.Done)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping crawler...")
			// Use context to ensure we don't block indefinitely during shutdown
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				p.Sources.Stop()
				return nil
			}
		},
	})

	return nil
}

func init() {
	rootCmd.AddCommand(crawlCmd)
}
