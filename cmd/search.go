package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search indexed content",
	Long:  `Search content that has been crawled and indexed in Elasticsearch`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		results, err := app.SearchContent(ctx, args[0], indexName, querySize)
		if err != nil {
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
	searchCmd.Flags().StringVarP(&indexName, "index", "i", "articles", "Index to search")
	searchCmd.Flags().IntVarP(&querySize, "size", "s", 10, "Number of results to return")
}
