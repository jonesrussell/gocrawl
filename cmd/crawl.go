package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
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
func NewCrawlCmd(log logger.Interface, cfg *config.Config, esClient *elasticsearch.Client) *cobra.Command {
	var crawlCmd = &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling a website",
		Long:  `Crawl a website and store the content in Elasticsearch`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupCrawlCmd(cmd, cfg)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeCrawlCmd(cmd, log, cfg, esClient)
		},
	}

	crawlCmd.Flags().StringP("url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntP("depth", "d", DefaultMaxDepth, "Maximum crawl depth")
	crawlCmd.Flags().DurationP("rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringP("index", "i", "articles", "Elasticsearch index name")

	err := crawlCmd.MarkFlagRequired("url")
	if err != nil {
		log.Error("Error marking URL flag as required", "error", err)
	}

	return crawlCmd
}

// setupCrawlCmd handles the setup for the crawl command
func setupCrawlCmd(cmd *cobra.Command, cfg *config.Config) error {
	// Set the configuration values directly from cfg
	cfg.Crawler.BaseURL = cmd.Flag("url").Value.String()
	if depth, err := cmd.Flags().GetInt("depth"); err == nil {
		cfg.Crawler.MaxDepth = depth
	}
	rateLimit, err := cmd.Flags().GetDuration("rate")
	if err == nil {
		cfg.Crawler.RateLimit = rateLimit
	}
	cfg.Crawler.IndexName = cmd.Flag("index").Value.String()
	return nil
}

// executeCrawlCmd handles the execution of the crawl command
func executeCrawlCmd(cmd *cobra.Command, log logger.Interface, cfg *config.Config, esClient *elasticsearch.Client) error {
	// Initialize fx container
	fxApp := fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,
		fx.Invoke(func(lc fx.Lifecycle, deps struct {
			fx.In
			Logger  logger.Interface
			Crawler *crawler.Crawler
		}) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Debug("Starting application...")
					return deps.Crawler.Start(ctx)
				},
				OnStop: func(ctx context.Context) error {
					log.Debug("Stopping application...")
					deps.Crawler.Stop()
					return nil
				},
			})
		}),
	)

	// Start the application
	if err := fxApp.Start(cmd.Context()); err != nil {
		log.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}
	defer func() {
		if err := fxApp.Stop(cmd.Context()); err != nil {
			log.Error("Error stopping application", "error", err)
		}
	}()

	// Use cfg here if needed, for example, to log the configuration
	log.Debug(fmt.Sprintf("Crawling with configuration: %+v", cfg.Crawler))

	// Use esClient as needed, for example, to log the configuration
	log.Debug(fmt.Sprintf("Crawling with Elasticsearch client: %+v", esClient))

	return nil
}
