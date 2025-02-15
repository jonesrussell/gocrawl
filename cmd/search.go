package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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
			// Retrieve the query from the -q flag
			query, err := cmd.Flags().GetString("query")
			if err != nil || query == "" {
				return fmt.Errorf("missing query argument")
			}

			// Initialize fx container
			var deps struct {
				fx.In
				Logger  logger.Interface
				Storage storage.Interface
			}

			fxApp := fx.New(
				config.Module,
				logger.Module,
				storage.Module,
				fx.Populate(&deps),
			)

			// Start the application
			if err := fxApp.Start(cmd.Context()); err != nil {
				log.Error("Error starting application", "error", err)
				return fmt.Errorf("error starting application: %w", err)
			}
			defer fxApp.Stop(cmd.Context())

			ctx := context.Background()

			// Use the injected storage instance to call SearchContent
			results, err := app.SearchContent(ctx, query, "articles", DefaultSearchSize) // Adjust parameters as needed
			if err != nil {
				log.Error("Search failed", err)
				return fmt.Errorf("search failed: %w", err)
			}

			// Print results using logger
			for _, result := range results {
				log.Info(fmt.Sprintf("URL: %s\nContent: %s\n\n", result.URL, result.Content))
			}

			return nil
		},
	}

	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	return searchCmd
}
