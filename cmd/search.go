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

// Constants for default values
const (
	DefaultSearchSize = 10 // Default number of results to return
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

			// Print results using logger instead of fmt.Printf
			for _, result := range results {
				log.Info(fmt.Sprintf("URL: %s\nContent: %s\n\n", result.URL, result.Content)) // Use logger
			}
			return nil
		},
	}

	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return") // Use the constant here

	return searchCmd
}
