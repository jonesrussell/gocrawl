package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultMaxDepth = 2 // Default maximum crawl depth
)

// NewCrawlCmd creates a new crawl command
func NewCrawlCmd(log logger.Interface) *cobra.Command {
	var crawlCmd = &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling a website",
		Long:  `Crawl a website and store the content in Elasticsearch`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupCrawlCmd(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeCrawlCmd(cmd, log)
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
func setupCrawlCmd(cmd *cobra.Command) error {
	viper.Set("CRAWLER_BASE_URL", cmd.Flag("url").Value.String())
	if depth, err := cmd.Flags().GetInt("depth"); err == nil {
		viper.Set("CRAWLER_MAX_DEPTH", depth)
	}
	viper.Set("CRAWLER_RATE_LIMIT", cmd.Flag("rate").Value.String())
	viper.Set("ELASTIC_INDEX_NAME", cmd.Flag("index").Value.String())
	return nil
}

// executeCrawlCmd handles the execution of the crawl command
func executeCrawlCmd(cmd *cobra.Command, log logger.Interface) error {
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

	return nil
}
