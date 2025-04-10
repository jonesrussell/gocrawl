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

// TableRenderer handles the display of index data in a table format
type TableRenderer struct {
	logger logger.Interface
}

// NewTableRenderer creates a new TableRenderer instance
func NewTableRenderer(logger logger.Interface) *TableRenderer {
	return &TableRenderer{
		logger: logger,
	}
}

// handleIndexError handles common error cases for index operations
func (r *TableRenderer) handleIndexError(operation, index string, err error, action, details string) error {
	r.logger.Error(fmt.Sprintf("Failed to %s for index", operation),
		"index", index,
		"error", err,
		"action", action,
		"details", details,
	)
	return fmt.Errorf("failed to %s for index %s: %w. %s", operation, index, err, action)
}

// RenderTable formats and displays the indices in a table format
func (r *TableRenderer) RenderTable(ctx context.Context, storage types.Interface, indices []string) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Status"})

	// Process each index and gather its metadata
	for _, index := range indices {
		// Get health status with error handling
		healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			return r.handleIndexError(
				"get health status",
				index,
				healthErr,
				"Check if the index exists and Elasticsearch is running",
				"This could be due to network issues, index not existing, or Elasticsearch being down",
			)
		}

		// Get document count with fallback to 0 on error
		docCount, docErr := storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			return r.handleIndexError(
				"get document count",
				index,
				docErr,
				"Check if the index exists and has documents",
				"This could be due to index corruption, permission issues, or Elasticsearch being in a degraded state",
			)
		}

		// Map health status to ingestion status
		ingestionStatus := getIngestionStatus(healthStatus)

		// Add row to table
		t.AppendRow([]any{
			index,
			healthStatus,
			docCount,
			ingestionStatus,
		})
	}

	// Render the final table
	t.Render()
	return nil
}

// Lister implements the indices list command
type Lister struct {
	config   config.Interface
	logger   logger.Interface
	storage  types.Interface
	renderer *TableRenderer
}

// NewLister creates a new lister instance
func NewLister(
	config config.Interface,
	logger logger.Interface,
	storage types.Interface,
	renderer *TableRenderer,
) *Lister {
	return &Lister{
		config:   config,
		logger:   logger,
		storage:  storage,
		renderer: renderer,
	}
}

// Start executes the list operation
func (l *Lister) Start(ctx context.Context) error {
	l.logger.Info("Listing indices")

	// Test storage connection
	if err := l.storage.TestConnection(ctx); err != nil {
		return fmt.Errorf(
			"failed to connect to storage: %w. Check Elasticsearch connection settings "+
				"and ensure the service is running",
			err,
		)
	}

	// List indices
	indices, err := l.storage.ListIndices(ctx)
	if err != nil {
		l.logger.Error("Failed to list indices",
			"error", err,
			"action", "Check Elasticsearch permissions and cluster health",
			"details", "This could be due to permission issues, cluster being in a degraded state, or network problems",
		)
		return fmt.Errorf("failed to list indices: %w. Check Elasticsearch permissions and cluster health", err)
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
		l.logger.Info("No indices found",
			"action", "Create an index using 'gocrawl indices create <index-name>'",
			"details", "No user-created indices were found in the Elasticsearch cluster",
		)
		return nil
	}

	// Render the indices table using the renderer
	return l.renderer.RenderTable(ctx, l.storage, filteredIndices)
}

// NewListCommand creates a new list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config path from flags
			configPath, _ := cmd.Flags().GetString("config")

			// Create Fx application
			app := fx.New(
				// Provide config path string
				fx.Provide(func() string { return configPath }),
				// Use the indices module
				Module,
				// Invoke list command
				fx.Invoke(func(l *Lister) error {
					return l.Start(cmd.Context())
				}),
			)

			// Start application
			if err := app.Start(context.Background()); err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			// Stop application
			if err := app.Stop(context.Background()); err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("config", "c", "config.yaml", "Path to config file")

	return cmd
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
