package cmd

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
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

// NewSearchCmd creates a new search command
func NewSearchCmd(log logger.Interface, cfg *config.Config, esClient *elasticsearch.Client) *cobra.Command {
	var searchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search content in Elasticsearch",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			// Set the ELASTIC_INDEX_NAME from the command flag
			cfg.Elasticsearch.IndexName = cmd.Flag("index").Value.String()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeSearchCmd(cmd, log, cfg, esClient)
		},
	}

	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	return searchCmd
}

// executeSearchCmd handles the search command execution
func executeSearchCmd(cmd *cobra.Command, log logger.Interface, cfg *config.Config, esClient *elasticsearch.Client) error {
	query, err := cmd.Flags().GetString("query")
	if err != nil || query == "" {
		return fmt.Errorf("missing query argument")
	}

	// Initialize fx container
	fxApp := fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		fx.Invoke(func(lc fx.Lifecycle, deps struct {
			fx.In
			Logger   logger.Interface
			Storage  storage.Interface
			ESClient *elasticsearch.Client
		}) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return runApp(ctx, deps.Logger, cfg, query, esClient)
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the application
	if err := fxApp.Start(cmd.Context()); err != nil {
		log.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}
	defer fxApp.Stop(cmd.Context())

	return nil
}

// runApp executes the main logic of the application
func runApp(ctx context.Context, log logger.Interface, cfg *config.Config, query string, esClient *elasticsearch.Client) error {
	results, err := app.SearchContent(ctx, esClient, query, cfg.Elasticsearch.IndexName, DefaultSearchSize)
	if err != nil {
		log.Error("Search failed", err)
		return fmt.Errorf("search failed: %w", err)
	}

	// Print results using logger
	for _, result := range results {
		log.Info(fmt.Sprintf("URL: %s\nContent: %s\n\n", result.URL, result.Content))
	}

	return nil
}
