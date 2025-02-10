package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/spf13/cobra"
)

var (
	query     string
	indexName string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search crawled content",
	Long:  `Search through the content stored in Elasticsearch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Convert query string to Elasticsearch query
		queryMap := map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"title", "content", "url"},
				},
			},
		}

		results, err := app.Search(ctx, indexName, queryMap)
		if err != nil {
			return err
		}

		// Pretty print results
		for _, result := range results {
			prettyJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(prettyJSON))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&query, "query", "q", "", "Search query (required)")
	searchCmd.Flags().StringVarP(&indexName, "index", "i", "pages", "Index name to search")

	searchCmd.MarkFlagRequired("query")
}
