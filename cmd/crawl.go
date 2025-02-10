package cmd

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/spf13/cobra"
)

var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Start crawling a website",
	Long:  `Crawl a website and store the content in Elasticsearch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return app.StartCrawler(ctx, &config.Config{
			CrawlerConfig: config.CrawlerConfig{
				BaseURL:   baseURL,
				MaxDepth:  maxDepth,
				RateLimit: rateLimit,
			},
			IndexName: indexName,
		})
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	crawlCmd.Flags().StringVarP(&baseURL, "url", "u", "", "Base URL to crawl (required)")
	crawlCmd.Flags().IntVarP(&maxDepth, "depth", "d", 2, "Maximum crawl depth")
	crawlCmd.Flags().DurationVarP(&rateLimit, "rate", "r", time.Second, "Rate limit between requests")
	crawlCmd.Flags().StringVarP(&indexName, "index", "i", "articles", "Elasticsearch index name")

	crawlCmd.MarkFlagRequired("url")
}
