package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultSearchSize = 10 // Default number of results to return
)

// SearchParams holds the parameters for the search operation
type SearchParams struct {
	fx.In

	Logger     common.Logger
	Config     common.Config
	SearchSvc  *search.Service
	IndexName  string `name:"indexName"`
	Query      string `name:"query"`
	ResultSize int    `name:"resultSize"`
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search content in Elasticsearch",
	RunE: func(cmd *cobra.Command, _ []string) error {
		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return fmt.Errorf("error retrieving query: %w", err)
		}

		indexName := cmd.Flag("index").Value.String()
		size, err := cmd.Flags().GetInt("size")
		if err != nil {
			size = DefaultSearchSize
		}

		app := fx.New(
			common.Module,
			search.Module,
			fx.Provide(
				// Provide search parameters
				fx.Annotate(
					func() string { return query },
					fx.ResultTags(`name:"query"`),
				),
				fx.Annotate(
					func() string { return indexName },
					fx.ResultTags(`name:"indexName"`),
				),
				fx.Annotate(
					func() int { return size },
					fx.ResultTags(`name:"resultSize"`),
				),
			),
			fx.Invoke(startSearch),
		)

		if err := app.Start(cmd.Context()); err != nil {
			fmt.Printf("Error starting application: %v\n", err)
			os.Exit(1)
		}

		// Wait for termination signal or search completion
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, initiating shutdown...\n", sig)
		case <-cmd.Context().Done():
			fmt.Println("\nSearch completed, shutting down...")
		}

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.Stop(ctx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

// startSearch initializes and runs the search operation
func startSearch(lc fx.Lifecycle, p SearchParams) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Logger.Info("Starting search...",
				"query", p.Query,
				"index", p.IndexName,
				"size", p.ResultSize,
			)

			results, err := p.SearchSvc.SearchContent(ctx, p.Query, p.IndexName, p.ResultSize)
			if err != nil {
				p.Logger.Error("Search failed", "error", err)
				return fmt.Errorf("search failed: %w", err)
			}

			// Print results
			if len(results) == 0 {
				p.Logger.Info("No results found")
				return nil
			}

			p.Logger.Info(fmt.Sprintf("Found %d results:", len(results)))
			for i, result := range results {
				p.Logger.Info(fmt.Sprintf("\nResult %d:", i+1),
					"url", result.URL,
					"content", result.Content,
				)
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			p.Logger.Debug("Search operation completed")
			return nil
		},
	})
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Define flags for the search command
	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	if err := searchCmd.MarkFlagRequired("query"); err != nil {
		fmt.Printf("Error marking query flag as required: %v\n", err)
		os.Exit(1)
	}
}
