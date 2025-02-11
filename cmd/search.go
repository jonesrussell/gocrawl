package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
)

// Modify searchCmd to accept a logger instance
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search indexed content",
	Long:  `Search content that has been crawled and indexed in Elasticsearch`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize the logger
		lgr, err := logger.NewLogger(&config.Config{App: config.AppConfig{Environment: os.Getenv("APP_ENV")}})
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Create config with CLI values
		cfg := &config.Config{
			Crawler: config.CrawlerConfig{
				IndexName: cmd.Flag("index").Value.String(),
				Transport: http.DefaultTransport,
			},
		}

		// Parse size parameter
		sizeStr := cmd.Flag("size").Value.String()
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			lgr.Error("Invalid size value", err)
			return fmt.Errorf("invalid size value: %w", err)
		}

		results, err := app.SearchContent(ctx, args[0], cfg.Crawler.IndexName, size)
		if err != nil {
			lgr.Error("Search failed", err)
			return fmt.Errorf("search failed: %w", err)
		}

		// Print results
		for _, result := range results {
			fmt.Printf("URL: %s\nContent: %s\n\n", result.URL, result.Content)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", 10, "Number of results to return")
}
