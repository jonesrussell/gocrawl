package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultSearchSize = 10 // Default number of results to return
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search content in Elasticsearch",
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return setupSearchCmd(cmd)
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		return executeSearchCmd(cmd)
	},
}

// setupSearchCmd handles the setup for the search command
func setupSearchCmd(cmd *cobra.Command) error {
	if globalConfig == nil {
		return errors.New("configuration is required") // Check if cfg is nil
	}

	// Use the SetIndexName method to set the index name
	indexName := cmd.Flag("index").Value.String()
	globalConfig.Crawler.SetIndexName(indexName) // Set the index name using the method

	return nil
}

// executeSearchCmd handles the search command execution
func executeSearchCmd(cmd *cobra.Command) error {
	query, err := cmd.Flags().GetString("query")
	if err != nil {
		globalLogger.Error("Error retrieving query", "error", err)
		return fmt.Errorf("error retrieving query: %w", err)
	}

	// Initialize fx container
	app := newSearchFxApp(query)

	// Start the application
	if startErr := app.Start(cmd.Context()); startErr != nil {
		globalLogger.Error("Error starting application", "error", startErr)
		return fmt.Errorf("error starting application: %w", startErr)
	}
	defer func() {
		if stopErr := app.Stop(cmd.Context()); stopErr != nil {
			globalLogger.Error("Error stopping application", "error", stopErr)
		}
	}()

	return nil
}

// newSearchFxApp initializes the Fx application with dependencies
func newSearchFxApp(query string) *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		search.Module,
		fx.Invoke(setupSearchLifecycleHooks),
		fx.Provide(func() string {
			return query
		}),
	)
}

// setupSearchLifecycleHooks sets up the lifecycle hooks for the Fx application
func setupSearchLifecycleHooks(lc fx.Lifecycle, deps struct {
	fx.In
	Logger    logger.Interface
	SearchSvc *search.Service
	Query     string
}) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Debug("Starting application...")
			return runSearchApp(ctx, deps.Logger, deps.SearchSvc, deps.Query)
		},
		OnStop: func(_ context.Context) error {
			deps.Logger.Debug("Stopping application...")
			return nil
		},
	})
}

// runSearchApp executes the main logic of the search application
func runSearchApp(ctx context.Context, log logger.Interface, searchSvc *search.Service, query string) error {
	// Use the index name from the global configuration
	indexName := globalConfig.Elasticsearch.IndexName

	results, err := searchSvc.SearchContent(ctx, query, indexName, DefaultSearchSize)
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

func init() {
	rootCmd.AddCommand(searchCmd)

	// Define flags for the search command in init
	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	err := searchCmd.MarkFlagRequired("query")
	if err != nil {
		globalLogger.Error("Error marking query flag as required", "error", err)
	}
}
