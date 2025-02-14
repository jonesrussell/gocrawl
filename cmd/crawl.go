package cmd

import (
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set viper values before config is created
			viper.Set("CRAWLER_BASE_URL", cmd.Flag("url").Value.String())
			viper.Set("CRAWLER_MAX_DEPTH", cmd.Flag("depth").Value.String())
			viper.Set("CRAWLER_RATE_LIMIT", cmd.Flag("rate").Value.String())
			viper.Set("ELASTIC_INDEX_NAME", cmd.Flag("index").Value.String())

			// Let fx lifecycle handle the crawler start/stop
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
