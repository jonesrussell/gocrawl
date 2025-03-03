package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultMaxDepth = 2 // Default maximum crawl depth
)

// CrawlParams holds the parameters for the crawl command
type CrawlParams struct {
	fx.In

	CrawlerInstance crawler.Interface
	Sources         *sources.Sources
	Processors      []models.ContentProcessor `group:"processors"`
}

// NewCrawlCmd creates a new crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Start crawling a website",
	Long:  `Crawl a website and store the content in Elasticsearch`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return setupCrawlCmd(cmd)
	},
	RunE: runCrawlCmd,
}

// setupCrawlCmd handles the setup for the crawl command
func setupCrawlCmd(cmd *cobra.Command) error {
	// Set the configuration values directly from cmd flags
	globalConfig.Crawler.BaseURL = cmd.Flag("url").Value.String()
	if depth, err := cmd.Flags().GetInt("depth"); err == nil {
		globalConfig.Crawler.MaxDepth = depth
	}
	rateLimit, err := cmd.Flags().GetDuration("rate")
	if err == nil {
		globalConfig.Crawler.RateLimit = rateLimit
	}
	globalConfig.Crawler.IndexName = cmd.Flag("index").Value.String()
	return nil
}

// runCrawlCmd handles the execution of the crawl command
func runCrawlCmd(cmd *cobra.Command, _ []string) error {
	// Initialize fx container
	app := fx.New(
		fx.Provide(
			func() *config.Config {
				return globalConfig // Provide the global config
			},
			func() logger.Interface {
				return globalLogger // Provide the global logger
			},
			// Provide article processor
			fx.Annotate(
				func() models.ContentProcessor {
					return &article.Processor{}
				},
				fx.ResultTags(`group:"processors"`),
			),
			// Provide content processor with dependencies
			fx.Annotate(
				func(
					service content.Interface,
					storage storage.Interface,
					logger logger.Interface,
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
		storage.Module,
		collector.Module,
		crawler.Module,
		sources.Module,
		content.Module,
		fx.Invoke(startCrawlCmd),
	)

	// Start the application
	if err := app.Start(cmd.Context()); err != nil {
		globalLogger.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}

	defer func() {
		if err := app.Stop(cmd.Context()); err != nil {
			globalLogger.Error("Error stopping application", "error", err)
		}
	}()

	return nil
}

// startCrawlCmd starts the crawl command
func startCrawlCmd(p CrawlParams) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Get the crawler instance to access index service
	crawler, ok := p.CrawlerInstance.(*crawler.Crawler)
	if !ok {
		return fmt.Errorf("crawler instance is not of type *crawler.Crawler")
	}

	// Set both the crawler and index manager in the sources
	p.Sources.SetCrawler(p.CrawlerInstance)
	p.Sources.SetIndexManager(crawler.IndexService)

	source, err := p.Sources.FindByName(sourceName)
	if err != nil {
		return err
	}

	// Parse rate limit
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	// Validate processors
	if len(p.Processors) < 2 {
		return fmt.Errorf("insufficient processors: need at least 2 processors (article and content), got %d", len(p.Processors))
	}

	// Create the collector using the collector module
	collectorResult, err := collector.New(collector.Params{
		BaseURL:          source.URL,
		MaxDepth:         source.MaxDepth,
		RateLimit:        rateLimit,
		Debugger:         logger.NewCollyDebugger(globalLogger),
		Logger:           globalLogger,
		ArticleProcessor: p.Processors[1], // Use second processor as article processor
		ContentProcessor: p.Processors[0], // Use first processor as content processor
		Source:           source,
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	crawler.SetCollector(collectorResult.Collector)

	// Start the crawl
	return p.Sources.Start(context.Background(), sourceName)
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	crawlCmd.Flags().StringP("url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntP("depth", "d", DefaultMaxDepth, "Maximum crawl depth")
	crawlCmd.Flags().DurationP("rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringP("index", "i", "articles", "Elasticsearch index name")

	err := crawlCmd.MarkFlagRequired("url")
	if err != nil {
		globalLogger.Error("Error marking URL flag as required", "error", err)
	}
}
