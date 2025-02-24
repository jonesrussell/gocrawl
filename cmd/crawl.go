package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultMaxDepth = 2 // Default maximum crawl depth
)

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
		),
		storage.Module,
		collector.Module,
		crawler.Module,
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

// startCrawl starts the crawling process
func startCrawlCmd(crawlerInstance *crawler.Crawler, articleProcessor *article.Processor) error {
	globalLogger.Debug("Starting crawl...")

	// Create the collector using global configuration
	params := collector.Params{
		BaseURL:          globalConfig.Crawler.BaseURL,
		MaxDepth:         globalConfig.Crawler.MaxDepth,
		RateLimit:        globalConfig.Crawler.RateLimit,
		Logger:           globalLogger,
		ArticleProcessor: articleProcessor,
	}

	// Create the collector
	collectorResult, err := collector.New(params)
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	crawlerInstance.SetCollector(collectorResult.Collector)

	// Start the crawling process
	if startErr := crawlerInstance.Start(context.Background(), params.BaseURL); startErr != nil {
		return fmt.Errorf("error starting crawler: %w", startErr)
	}

	return nil
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
