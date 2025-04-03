// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// ListDeps holds the dependencies for the list command.
type ListDeps struct {
	fx.In

	Context      context.Context `name:"indicesContext"`
	Config       config.Interface
	Logger       logger.Interface
	IndexManager api.IndexManager
}

// Dependencies holds the list command's dependencies
type Dependencies struct {
	fx.In

	Lifecycle fx.Lifecycle
	Storage   storagetypes.Interface
	Logger    logger.Interface
	Context   context.Context `name:"crawlContext"`
}

// listModule provides the list command dependencies
var listModule = fx.Module("list",
	// Core dependencies
	config.Module,
	storage.Module,
)

// listCommand creates and returns the list command that displays all indices.
// It integrates with the Cobra command framework and sets up the command
// structure with appropriate usage information and examples.
//
// Returns:
//   - *cobra.Command: A configured Cobra command ready to be added to the root command
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		RunE: runList,
	}
}

// renderIndicesTable formats and displays the indices in a table format.
func renderIndicesTable(
	ctx context.Context,
	indices []string,
	storage storagetypes.Interface,
	logger logger.Interface,
) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Ingestion Name", "Ingestion Status"})

	// Process each index and gather its metadata
	for _, index := range indices {
		// Get health status with error handling
		healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			logger.Error("Error getting health for index", "index", index, "error", healthErr)
			continue
		}

		// Get document count with fallback to 0 on error
		docCount, docErr := storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			docCount = 0
		}

		// Map health status to ingestion status
		ingestionStatus := getIngestionStatus(healthStatus)

		// Add row to table
		t.AppendRow([]any{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for future ingestion name feature
			ingestionStatus,
		})
	}

	// Render the final table
	t.Render()
	return nil
}

// runList executes the list command and displays all indices.
// It:
// - Initializes the Fx application with required modules
// - Handles application lifecycle and error cases
// - Displays the indices list in a formatted table
func runList(cmd *cobra.Command, _ []string) error {
	// Create a context
	ctx := cmd.Context()

	// Initialize the Fx application
	fxApp := fx.New(
		fx.NopLogger,
		listModule,
		fx.Provide(
			fx.Annotate(
				func() context.Context { return ctx },
				fx.ResultTags(`name:"crawlContext"`),
			),
			logger.NewNoOp,
		),
		fx.Invoke(func(lc fx.Lifecycle, p Dependencies) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := p.Storage.TestConnection(ctx); err != nil {
						return fmt.Errorf("failed to connect to storage: %w", err)
					}

					// List indices
					indices, err := p.Storage.ListIndices(ctx)
					if err != nil {
						return fmt.Errorf("failed to list indices: %w", err)
					}

					// Filter out internal indices (those starting with '.')
					var filteredIndices []string
					for _, index := range indices {
						if !strings.HasPrefix(index, ".") {
							filteredIndices = append(filteredIndices, index)
						}
					}

					// Handle the case where no indices are found
					if len(filteredIndices) == 0 {
						p.Logger.Info("No indices found")
						return nil
					}

					// Render the indices table
					return renderIndicesTable(ctx, filteredIndices, p.Storage, p.Logger)
				},
				OnStop: func(ctx context.Context) error {
					p.Logger.Info("Stopping...")
					return nil
				},
			})
		}),
	)

	// Start the application
	if err := fxApp.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Stop the application immediately after starting
	if err := fxApp.Stop(ctx); err != nil {
		return fmt.Errorf("error stopping application: %w", err)
	}

	return nil
}

// getIngestionStatus maps Elasticsearch health status to human-readable ingestion status.
// This function provides a user-friendly interpretation of index health:
// - "red" indicates a serious issue (Disconnected)
// - "yellow" indicates a potential issue (Warning)
// - Any other status is considered healthy (Connected)
//
// Parameters:
//   - healthStatus: The Elasticsearch health status string
//
// Returns:
//   - string: A human-readable ingestion status
func getIngestionStatus(healthStatus string) string {
	switch healthStatus {
	case "red":
		return "Disconnected"
	case "yellow":
		return "Warning"
	default:
		return "Connected"
	}
}
