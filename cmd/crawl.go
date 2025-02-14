package cmd

import (
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

// NewCrawlCmd creates a new crawl command
func NewCrawlCmd(log logger.Interface, cfg *config.Config, storageInstance storage.Interface) *cobra.Command {
	var crawlCmd = &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling a website",
		Long:  `Crawl a website and store the content in Elasticsearch`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			viper.Set("CRAWLER_BASE_URL", cmd.Flag("url").Value.String())
			if depth, err := cmd.Flags().GetInt("depth"); err == nil {
				viper.Set("CRAWLER_MAX_DEPTH", depth)
			}
			viper.Set("CRAWLER_RATE_LIMIT", cmd.Flag("rate").Value.String())
			viper.Set("ELASTIC_INDEX_NAME", cmd.Flag("index").Value.String())
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Initialize fx container
			var deps struct {
				fx.In
				Logger logger.Interface
			}

			app := fx.New(
				config.Module,
				logger.Module,
				storage.Module,
				collector.Module,
				crawler.Module,
				fx.Populate(&deps),
			)

			// Debug log before starting the application
			log.Debug("Starting application...")

			if err := app.Start(cmd.Context()); err != nil {
				log.Error("Error starting application", "error", err)
				return fmt.Errorf("error starting application: %w", err)
			}
			defer app.Stop(cmd.Context())

			// Debug log after successful start
			log.Debug("Application started successfully")
			return nil
		},
	}

	crawlCmd.Flags().StringP("url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntP("depth", "d", 2, "Maximum crawl depth")
	crawlCmd.Flags().DurationP("rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringP("index", "i", "articles", "Elasticsearch index name")

	err := crawlCmd.MarkFlagRequired("url")
	if err != nil {
		log.Error("Error marking URL flag as required", "error", err)
	}

	return crawlCmd
}
