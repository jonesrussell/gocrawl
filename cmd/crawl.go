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
	"github.com/jonesrussell/gocrawl/internal/config"
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

// SetSourceName sets the source name for testing purposes
func SetSourceName(name string) {
	sourceName = name
}

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
	Logger logger.Interface

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`

	// Config holds the application configuration
	Config *config.Config

	// Context provides the context for the crawl operation
	Context context.Context `name:"crawlContext"`
}

// CrawlCmd represents the crawl command that initiates the crawling process for a
// specified source. It reads the source configuration from sources.yml and starts
// the crawling process with the configured parameters.
var CrawlCmd = &cobra.Command{
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

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create channels for error handling and completion
		errChan := make(chan error, 1)
		doneChan := make(chan struct{})

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			crawler.Module,
			article.Module,
			content.Module,
			collector.Module(),
			fx.Provide(
				func() chan struct{} {
					return doneChan
				},
				fx.Annotate(
					func() chan struct{} {
						return doneChan
					},
					fx.ResultTags(`name:"crawlDone"`),
				),
				fx.Annotate(
					func() context.Context {
						return cmd.Context()
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
					func(sources *sources.Sources) string {
						// Get the source configuration
						source, err := sources.FindByName(sourceName)
						if err != nil {
							return "" // Return empty string if source not found
						}
						return source.Index
					},
					fx.ResultTags(`name:"contentIndex"`),
				),
				fx.Annotate(
					func(sources *sources.Sources) string {
						// Get the source configuration
						source, err := sources.FindByName(sourceName)
						if err != nil {
							return "" // Return empty string if source not found
						}
						return source.ArticleIndex
					},
					fx.ResultTags(`name:"indexName"`),
				),
				func() chan *models.Article {
					return make(chan *models.Article, DefaultChannelBufferSize)
				},
			),
			fx.Invoke(func(lc fx.Lifecycle, p CrawlParams) {
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						// Execute the crawl and handle any errors
						if err := startCrawl(p); err != nil {
							p.Logger.Error("Error executing crawl", "error", err)
							common.PrintErrorf("\nCrawl failed: %v", err)
							errChan <- err
							return err
						}
						close(doneChan)
						return nil
					},
					OnStop: func(context.Context) error {
						return nil
					},
				})
			}),
		)

		// Start the application and handle any startup errors
		if err := app.Start(cmd.Context()); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either:
		// - A signal interrupt (SIGINT/SIGTERM)
		// - Context cancellation
		// - Crawl completion
		// - Crawl error
		var crawlErr error
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
		case <-cmd.Context().Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		case crawlErr = <-errChan:
			// Error already printed in startCrawl
		case <-doneChan:
			// Success message already printed in startCrawl
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(cmd.Context(), common.DefaultOperationTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
			return err
		}

		return crawlErr
	},
}

// StartCrawl initializes and starts the crawling process. It sets up the crawler
// with the specified source configuration and manages the crawl lifecycle.
//
// Parameters:
//   - p: CrawlParams containing all required dependencies and configuration
//
// Returns:
//   - error: Any error that occurred during setup or initialization
func StartCrawl(p CrawlParams) error {
	return startCrawl(p)
}

// startCrawl is the internal implementation of StartCrawl
func startCrawl(p CrawlParams) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Configure the sources with crawler
	p.Sources.SetCrawler(p.CrawlerInstance)

	// Get the source configuration
	source, err := p.Sources.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("error finding source: %w", err)
	}

	// Convert the source config to the expected type
	sourceConfig := convertSourceConfig(source)
	if sourceConfig == nil {
		return errors.New("source configuration is nil")
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
		Source:           sourceConfig,
		Parallelism:      p.Config.Crawler.Parallelism,
		RandomDelay:      p.Config.Crawler.RandomDelay,
		Context:          p.Context,
		Done:             p.Done,
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	p.CrawlerInstance.SetCollector(collectorResult.Collector)

	// Set the IndexManager in the Sources struct
	p.Sources.SetIndexManager(p.CrawlerInstance.GetIndexManager())

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

// convertSourceConfig converts a sources.Config to a config.Source.
// It handles the conversion of fields between the two types.
func convertSourceConfig(source *sources.Config) *config.Source {
	if source == nil {
		return nil
	}

	// Parse the rate limit string into a duration
	rateLimit, err := config.ParseRateLimit(source.RateLimit)
	if err != nil {
		rateLimit = time.Second // Default to 1 second if parsing fails
	}

	return &config.Source{
		Name:         source.Name,
		URL:          source.URL,
		ArticleIndex: source.ArticleIndex,
		Index:        source.Index,
		RateLimit:    rateLimit,
		MaxDepth:     source.MaxDepth,
		Time:         source.Time,
		Selectors: config.SourceSelectors{
			Article: config.ArticleSelectors{
				Container:     source.Selectors.Article.Container,
				Title:         source.Selectors.Article.Title,
				Body:          source.Selectors.Article.Body,
				Intro:         source.Selectors.Article.Intro,
				Byline:        source.Selectors.Article.Byline,
				PublishedTime: source.Selectors.Article.PublishedTime,
				TimeAgo:       source.Selectors.Article.TimeAgo,
				JSONLD:        source.Selectors.Article.JSONLd,
				Section:       source.Selectors.Article.Section,
				Keywords:      source.Selectors.Article.Keywords,
				Description:   source.Selectors.Article.Description,
				OGTitle:       source.Selectors.Article.OgTitle,
				OGDescription: source.Selectors.Article.OgDescription,
				OGImage:       source.Selectors.Article.OgImage,
				OgURL:         source.Selectors.Article.OgURL,
				Canonical:     source.Selectors.Article.Canonical,
				WordCount:     source.Selectors.Article.WordCount,
				PublishDate:   source.Selectors.Article.PublishDate,
				Category:      source.Selectors.Article.Category,
				Tags:          source.Selectors.Article.Tags,
				Author:        source.Selectors.Article.Author,
				BylineName:    source.Selectors.Article.BylineName,
			},
		},
	}
}

func init() {
	rootCmd.AddCommand(CrawlCmd)
}
