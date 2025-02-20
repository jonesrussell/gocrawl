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
	RunE: func(cmd *cobra.Command, _ []string) error {
		return executeCrawlCmd(cmd)
	},
}

// setupCrawlCmd handles the setup for the crawl command
func setupCrawlCmd(cmd *cobra.Command) error {
	// Set the configuration values directly from cfg
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

// executeCrawlCmd handles the execution of the crawl command
func executeCrawlCmd(cmd *cobra.Command) error {
	// Initialize fx container
	app := newCrawlFxApp()

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

// newFxApp initializes the Fx application with dependencies
func newCrawlFxApp() *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,
		fx.Invoke(setupLifecycleHooks),
	)
}

// setupLifecycleHooks sets up the lifecycle hooks for the Fx application
func setupLifecycleHooks(lc fx.Lifecycle, deps struct {
	fx.In
	Logger  logger.Interface
	Crawler *crawler.Crawler
}) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Debug("Starting application...")
			return deps.Crawler.Start(ctx, deps.Crawler.GetBaseURL())
		},
		OnStop: func(_ context.Context) error {
			deps.Logger.Debug("Stopping application...")
			deps.Crawler.Stop()
			return nil
		},
	})
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
