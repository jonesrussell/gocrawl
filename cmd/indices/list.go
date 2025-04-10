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
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Lister implements the indices list command
type Lister struct {
	config  config.Interface
	logger  logger.Interface
	storage types.Interface
}

// NewLister creates a new lister instance
func NewLister(
	config config.Interface,
	logger logger.Interface,
	storage types.Interface,
) *Lister {
	return &Lister{
		config:  config,
		logger:  logger,
		storage: storage,
	}
}

// Start executes the list operation
func (l *Lister) Start(ctx context.Context) error {
	l.logger.Info("Listing indices")

	// Test storage connection
	if err := l.storage.TestConnection(ctx); err != nil {
		l.logger.Error("Failed to connect to storage", "error", err)
		return fmt.Errorf("failed to connect to storage: %w", err)
	}

	// List indices
	indices, err := l.storage.ListIndices(ctx)
	if err != nil {
		l.logger.Error("Failed to list indices", "error", err)
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
		l.logger.Info("No indices found")
		return nil
	}

	// Render the indices table
	return l.renderIndicesTable(ctx, filteredIndices)
}

// renderIndicesTable formats and displays the indices in a table format.
func (l *Lister) renderIndicesTable(ctx context.Context, indices []string) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Ingestion Name", "Ingestion Status"})

	// Process each index and gather its metadata
	for _, index := range indices {
		// Get health status with error handling
		healthStatus, healthErr := l.storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			l.logger.Error("Error getting health for index", "index", index, "error", healthErr)
			return fmt.Errorf("failed to get health for index %s: %w", index, healthErr)
		}

		// Get document count with fallback to 0 on error
		docCount, docErr := l.storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			l.logger.Error("Error getting document count for index", "index", index, "error", docErr)
			return fmt.Errorf("failed to get document count for index %s: %w", index, docErr)
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

// listCommand creates and returns the list command that displays all indices.
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create a context
			ctx := cmd.Context()

			// Get config path from flag or use default
			configPath, _ := cmd.Flags().GetString("config")

			// Initialize the Fx application
			fxApp := fx.New(
				fx.NopLogger,
				Module,
				fx.Provide(
					func() context.Context { return ctx },
					func() string { return configPath }, // Provide config path
				),
				fx.Invoke(func(lc fx.Lifecycle, lister *Lister) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							return lister.Start(ctx)
						},
						OnStop: func(context.Context) error {
							return nil
						},
					})
				}),
			)

			// Start the application
			if err := fxApp.Start(ctx); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			// Stop the application
			if err := fxApp.Stop(ctx); err != nil {
				return fmt.Errorf("error stopping application: %w", err)
			}

			return nil
		},
	}
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
