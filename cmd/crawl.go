package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func NewCrawlCmd(lgr *logger.CustomLogger) *cobra.Command {
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
			viper.Set("ELASTIC_INDEX_NAME", cmd.Flag("index").Value.String())

			// Create config with CLI values
			cfg := &config.Config{
				Crawler: config.CrawlerConfig{
					BaseURL:   cmd.Flag("url").Value.String(),
					IndexName: cmd.Flag("index").Value.String(),
					Transport: http.DefaultTransport,
				},
				Elasticsearch: config.ElasticsearchConfig{
					URL: os.Getenv("ELASTIC_URL"),
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

			// Get the crawler instance from the fx container
			var c *crawler.Crawler
			if err := fx.New(
				fx.Populate(&c),
				fx.Supply(cfg),
			).Start(ctx); err != nil {
				lgr.Error("Error getting crawler instance", err)
				return err
			}

			// Start the crawler
			if err := c.Start(ctx); err != nil {
				lgr.Error("Error starting crawler", err)
				return err
			}

			return nil
		},
	}

	crawlCmd.Flags().StringP("url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntP("depth", "d", 2, "Maximum crawl depth")
	crawlCmd.Flags().DurationP("rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringP("index", "i", "articles", "Elasticsearch index name")

	err := crawlCmd.MarkFlagRequired("url")
	if err != nil {
		return nil
	}

	return crawlCmd
}
