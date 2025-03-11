// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

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

// Params holds the dependencies required for the job scheduler.
type Params struct {
	fx.In

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// Sources provides access to configured crawl sources
	Sources *sources.Sources

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Logger provides structured logging capabilities
	Logger logger.Interface

	// Config holds the application configuration
	Config config.Interface

	// Context provides the context for the job scheduler
	Context context.Context `name:"jobContext"`

	// Processors is a slice of content processors, injected as a group
	Processors []models.ContentProcessor `group:"processors"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(
	ctx context.Context,
	log common.Logger,
	sources *sources.Sources,
	c crawler.Interface,
	processors []models.ContentProcessor,
	done chan struct{},
	cfg config.Interface,
) {
	log.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Do initial check
	checkAndRunJobs(ctx, log, sources, c, time.Now(), processors, done, cfg)

	for {
		select {
		case <-ctx.Done():
			log.Info("Job scheduler shutting down")
			return
		case t := <-ticker.C:
			checkAndRunJobs(ctx, log, sources, c, t, processors, done, cfg)
		}
	}
}

// setupCollector creates and configures a new collector for the given source.
func setupCollector(
	ctx context.Context,
	log common.Logger,
	source sources.Config,
	processors []models.ContentProcessor,
	done chan struct{},
	cfg config.Interface,
) (collector.Result, error) {
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		return collector.Result{}, fmt.Errorf("invalid rate limit format: %w", err)
	}

	// Convert source config to the expected type
	sourceConfig := convertSourceConfig(&source)
	if sourceConfig == nil {
		return collector.Result{}, errors.New("source configuration is nil")
	}

	// Extract domain from source URL
	parsedURL, err := url.Parse(source.URL)
	if err != nil {
		return collector.Result{}, fmt.Errorf("invalid source URL: %w", err)
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

	return collector.New(collector.Params{
		BaseURL:          source.URL,
		MaxDepth:         source.MaxDepth,
		RateLimit:        rateLimit,
		Logger:           log,
		Context:          ctx,
		ArticleProcessor: processors[0], // First processor handles articles
		ContentProcessor: processors[1], // Second processor handles content
		Done:             done,
		Debugger:         logger.NewCollyDebugger(log),
		Source:           sourceConfig,
		Parallelism:      cfg.GetCrawlerConfig().Parallelism,
		RandomDelay:      cfg.GetCrawlerConfig().RandomDelay,
		AllowedDomains:   []string{domain},
	})
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

// configureCrawler sets up the crawler with the given source configuration.
func configureCrawler(c crawler.Interface, source sources.Config, collector collector.Result) error {
	c.SetCollector(collector.Collector)
	c.SetMaxDepth(source.MaxDepth)
	if err := c.SetRateLimit(source.RateLimit); err != nil {
		return fmt.Errorf("error setting rate limit: %w", err)
	}
	return nil
}

// executeCrawl performs the crawl operation for a single source.
func executeCrawl(
	ctx context.Context,
	log common.Logger,
	c crawler.Interface,
	source sources.Config,
	processors []models.ContentProcessor,
	done chan struct{},
	cfg config.Interface,
) {
	collectorResult, err := setupCollector(ctx, log, source, processors, done, cfg)
	if err != nil {
		log.Error("Error setting up collector",
			"error", err,
			"source", source.Name)
		return
	}

	if configErr := configureCrawler(c, source, collectorResult); configErr != nil {
		log.Error("Error configuring crawler",
			"error", configErr,
			"source", source.Name)
		return
	}

	if startErr := c.Start(ctx, source.URL); startErr != nil {
		log.Error("Error starting crawler",
			"error", startErr,
			"source", source.Name)
		return
	}

	c.Wait()
	log.Info("Crawl completed", "source", source.Name)
}

// checkAndRunJobs evaluates and executes scheduled jobs.
func checkAndRunJobs(
	ctx context.Context,
	log common.Logger,
	sources *sources.Sources,
	c crawler.Interface,
	now time.Time,
	processors []models.ContentProcessor,
	done chan struct{},
	cfg config.Interface,
) {
	if sources == nil {
		log.Error("Sources configuration is nil")
		return
	}

	if c == nil {
		log.Error("Crawler instance is nil")
		return
	}

	currentTime := now.Format("15:04")
	log.Info("Checking jobs", "current_time", currentTime)

	for _, source := range sources.Sources {
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				log.Info("Running scheduled crawl",
					"source", source.Name,
					"time", scheduledTime)
				executeCrawl(ctx, log, c, source, processors, done, cfg)
			}
		}
	}
}

// startJob initializes and starts the job scheduler.
func startJob(p Params) error {
	// Create cancellable context
	ctx, cancel := context.WithCancel(p.Context)
	defer cancel()

	// Print loaded schedules
	p.Logger.Info("Loaded schedules:")
	for _, source := range p.Sources.Sources {
		if len(source.Time) > 0 {
			p.Logger.Info("Source schedule",
				"name", source.Name,
				"times", source.Time)
		}
	}

	// Start scheduler in background
	go runScheduler(ctx, p.Logger, p.Sources, p.CrawlerInstance, p.Processors, p.Done, p.Config)

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	p.Logger.Info("Job scheduler running. Press Ctrl+C to stop...")
	<-sigChan
	p.Logger.Info("Shutting down...")

	return nil
}

// provideSources creates a new Sources instance from the sources.yml file.
func provideSources() (*sources.Sources, error) {
	return sources.LoadFromFile("sources.yml")
}

// Cmd represents the job scheduler command.
var Cmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl jobs",
	Long:  `Schedule and run crawl jobs based on the times specified in sources.yml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Create a parent context that can be cancelled
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			article.Module,
			content.Module,
			collector.Module(),
			crawler.Module,
			fx.Provide(
				fx.Annotate(
					func() context.Context {
						return ctx
					},
					fx.ResultTags(`name:"jobContext"`),
				),
				provideSources,
				fx.Annotate(
					func(sources *sources.Sources) (string, string) {
						// For job scheduler, we'll use the first source's indices
						if len(sources.Sources) > 0 {
							return sources.Sources[0].Index, sources.Sources[0].ArticleIndex
						}
						return "content", "articles" // Default indices if no sources
					},
					fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
				),
				fx.Annotate(
					func(sources *sources.Sources) string {
						// For job scheduler, we'll use the first source's name
						if len(sources.Sources) > 0 {
							return sources.Sources[0].Name
						}
						return "default" // Default source name if no sources
					},
					fx.ResultTags(`name:"sourceName"`),
				),
				fx.Annotate(
					func() chan struct{} {
						return make(chan struct{})
					},
					fx.ResultTags(`name:"crawlDone"`),
				),
				func() chan *models.Article {
					return make(chan *models.Article, DefaultChannelBufferSize)
				},
			),
			fx.Invoke(startJob),
		)

		// Start the application and handle any startup errors
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either:
		// - A signal interrupt (SIGINT/SIGTERM)
		// - Context cancellation
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
			cancel() // Cancel our context
		case <-ctx.Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
			return err
		}

		return nil
	},
}

// Command returns the job command.
func Command() *cobra.Command {
	return Cmd
}
