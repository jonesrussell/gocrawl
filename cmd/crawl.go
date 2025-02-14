package cmd

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCrawlCmd(lgr *logger.CustomLogger, crawler *crawler.Crawler) *cobra.Command {
	var crawlCmd = &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling a website",
		Long:  `Crawl a website and store the content in Elasticsearch`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Set viper values before command runs
			viper.Set("CRAWLER_BASE_URL", cmd.Flag("url").Value.String())
			if depth, err := cmd.Flags().GetInt("depth"); err == nil {
				viper.Set("CRAWLER_MAX_DEPTH", depth)
			}
			viper.Set("CRAWLER_RATE_LIMIT", cmd.Flag("rate").Value.String())
			viper.Set("ELASTIC_INDEX_NAME", cmd.Flag("index").Value.String())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- crawler.Start(ctx)
			}()

			select {
			case err := <-done:
				if err != nil {
					lgr.Error("Crawler failed", err)
					return err
				}
				lgr.Info("Crawler completed successfully")
			case <-ctx.Done():
				lgr.Info("Crawler interrupted")
				return ctx.Err()
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
