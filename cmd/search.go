package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
)

func NewSearchCmd(log logger.Interface) *cobra.Command {
	var searchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search content in Elasticsearch",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Skip execution during flag parsing phase
			if log == nil {
				return nil
			}
			ctx := context.Background()

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
				log.Error("Invalid size value", err)
				return fmt.Errorf("invalid size value: %w", err)
			}

			results, err := app.SearchContent(ctx, args[0], cfg.Crawler.IndexName, size)
			if err != nil {
				log.Error("Search failed", err)
				return fmt.Errorf("search failed: %w", err)
			}

			// Print results
			for _, result := range results {
				fmt.Printf("URL: %s\nContent: %s\n\n", result.URL, result.Content)
			}
			return nil
		},
	}

	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", 10, "Number of results to return")

	return searchCmd
}
