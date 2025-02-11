package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Start crawling a website",
	Long:  `Crawl a website and store the content in Elasticsearch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Set viper values
		viper.Set("CRAWLER_BASE_URL", cmd.Flag("url").Value.String())
		viper.Set("CRAWLER_MAX_DEPTH", cmd.Flag("depth").Value.String())
		viper.Set("CRAWLER_RATE_LIMIT", cmd.Flag("rate").Value.String())
		viper.Set("INDEX_NAME", cmd.Flag("index").Value.String())

		// Create config with CLI values
		cfg := &config.Config{
			Crawler: config.CrawlerConfig{
				BaseURL:   cmd.Flag("url").Value.String(),
				IndexName: cmd.Flag("index").Value.String(),
				Transport: http.DefaultTransport,
			},
		}

		// Parse rate limit
		if rateStr := cmd.Flag("rate").Value.String(); rateStr != "" {
			rate, err := time.ParseDuration(rateStr)
			if err != nil {
				return fmt.Errorf("invalid rate value: %w", err)
			}
			cfg.Crawler.RateLimit = rate
		}

		// Parse max depth
		if depthStr := cmd.Flag("depth").Value.String(); depthStr != "" {
			depth, err := strconv.Atoi(depthStr)
			if err != nil {
				return fmt.Errorf("invalid depth value: %w", err)
			}
			cfg.Crawler.MaxDepth = depth
		}

		return app.StartCrawler(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	crawlCmd.Flags().StringP("url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntP("depth", "d", 2, "Maximum crawl depth")
	crawlCmd.Flags().DurationP("rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringP("index", "i", "articles", "Elasticsearch index name")

	crawlCmd.MarkFlagRequired("url")
}
