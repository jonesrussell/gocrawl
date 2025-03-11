// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
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

// Params holds the dependencies and parameters required for the crawl operation.
// It uses fx.In to indicate that these fields should be injected by the fx dependency
// injection framework.
type Params struct {
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
	Config config.Interface

	// Context provides the context for the crawl operation
	Context context.Context `name:"crawlContext"`
}

// Cmd represents the crawl command that initiates the crawling process for a
// specified source. It reads the source configuration from sources.yml and starts
// the crawling process with the configured parameters.
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

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create channels for error handling and completion
		errChan := make(chan error, 1)
		doneChan := make(chan struct{})

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			sources.Module,
			crawler.Module,
			article.Module,
			content.Module,
			collector.Module(),
			Module,
			fx.Provide(
				fx.Annotate(
					func() chan struct{} {
						return doneChan
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
					return make(chan *models.Article, DefaultChannelBufferSize)
				},
			),
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
				configureCrawlLifecycle(lc, p, errChan)
			}),
		)

		// Start the application and handle any startup errors
		if err := app.Start(ctx); err != nil {
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
			cancel() // Cancel our context
		case <-ctx.Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		case crawlErr = <-errChan:
			// Error already printed in startCrawl
			cancel() // Cancel our context on error
		case <-doneChan:
			// Success message already printed in startCrawl
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
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
//   - p: Params containing all required dependencies and configuration
//
// Returns:
//   - error: Any error that occurred during setup or initialization
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

	// Extract domain from source URL
	parsedURL, err := url.Parse(source.URL)
	if err != nil {
		return fmt.Errorf("invalid source URL: %w", err)
	}

	// Extract domain from URL, handling both full URLs and path-only URLs
	var domain string
	if parsedURL.Host == "" {
		// If no host in URL, treat the first path segment as the domain
		parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(parts) > 0 {
			domain = parts[0]
		}
	} else {
		// For full URLs, use the host as domain
		domain = parsedURL.Host
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
		Parallelism:      p.Config.GetCrawlerConfig().Parallelism,
		RandomDelay:      p.Config.GetCrawlerConfig().RandomDelay,
		Context:          p.Context,
		Done:             p.Done,
		AllowedDomains:   []string{domain},
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	p.CrawlerInstance.SetCollector(collectorResult.Collector)

	// Set crawler configuration
	p.CrawlerInstance.SetMaxDepth(source.MaxDepth)
	if rateLimitErr := p.CrawlerInstance.SetRateLimit(source.RateLimit); rateLimitErr != nil {
		return fmt.Errorf("error setting rate limit: %w", rateLimitErr)
	}

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

// initializeCrawl initializes the crawler with the given source
func initializeCrawl(_ context.Context, p Params) (*sources.Config, error) {
	source, findErr := p.Sources.FindByName(sourceName)
	if findErr != nil {
		return nil, fmt.Errorf("error finding source: %w", findErr)
	}

	if err := startCrawl(p); err != nil {
		return nil, fmt.Errorf("failed to initialize crawler: %w", err)
	}

	return source, nil
}

// runCrawl executes the crawl operation in a goroutine
func runCrawl(ctx context.Context, p Params, source *sources.Config, errChan chan error) {
	go func() {
		p.Logger.Info("Starting crawl", "source", sourceName)
		if startErr := p.CrawlerInstance.Start(ctx, source.URL); startErr != nil {
			if !errors.Is(startErr, context.Canceled) {
				p.Logger.Error("Crawl failed", "error", startErr)
				errChan <- startErr
			}
			close(p.Done)
			return
		}

		p.CrawlerInstance.Wait()
		p.Logger.Info("Crawl completed successfully", "source", sourceName)
		close(p.Done)
	}()
}

// handleShutdown manages the graceful shutdown of the crawler
func handleShutdown(ctx context.Context, p Params) error {
	p.Logger.Info("Stopping crawler...")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.Done:
		p.Logger.Info("Crawl completed, stopping gracefully")
		if stopErr := p.CrawlerInstance.Stop(ctx); stopErr != nil {
			p.Logger.Error("Error stopping crawler", "error", stopErr)
		}
		return nil
	default:
		p.Logger.Info("Forcing crawler to stop")
		if stopErr := p.CrawlerInstance.Stop(ctx); stopErr != nil {
			p.Logger.Error("Error stopping crawler", "error", stopErr)
		}
		return nil
	}
}

// configureCrawlLifecycle configures the lifecycle hooks for crawl management.
// It sets up the start and stop hooks for the crawl operation.
func configureCrawlLifecycle(lc fx.Lifecycle, p Params, errChan chan error) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			source, err := initializeCrawl(ctx, p)
			if err != nil {
				return err
			}

			runCrawl(ctx, p, source, errChan)

			// Wait for either error or completion
			select {
			case channelErr := <-errChan:
				return channelErr
			case <-p.Done:
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			return handleShutdown(ctx, p)
		},
	})
}

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
