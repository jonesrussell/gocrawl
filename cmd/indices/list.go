// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for table formatting
const (
	// TableWidth defines the total width of the table output for consistent formatting.
	// This ensures that the output remains readable across different terminal sizes.
	TableWidth = 92
)

// listParams holds the parameters required for listing indices.
// It uses fx.In to enable dependency injection of required components.
// This struct is used internally by the executeList function to manage
// dependencies and maintain clean separation of concerns.
type listParams struct {
	fx.In

	// ctx is the context for managing timeouts and cancellation
	ctx context.Context

	// storage provides access to Elasticsearch operations
	storage common.Storage

	// logger is used for structured logging throughout the command
	logger common.Logger
}

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
		Run: runList,
	}
}

// runList executes the list command and displays all indices.
// This function serves as the main entry point for the list command and handles:
// - Application lifecycle using Fx dependency injection
// - Context management for graceful shutdown
// - Error handling and logging
// - Exit code management
//
// Parameters:
//   - cmd: The Cobra command instance providing command context
//   - _: Unused args parameter
//
// The function uses a deferred shutdown sequence to ensure proper cleanup
// regardless of how the command exits.
func runList(cmd *cobra.Command, _ []string) {
	// Initialize variables for logger and exit code management
	var logger common.Logger
	var exitCode int

	// Initialize the Fx application with required modules
	// This sets up dependency injection and manages the application lifecycle
	app := fx.New(
		common.Module,
		fx.Invoke(func(s common.Storage, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				storage: s,
				logger:  l,
			}
			if err := executeList(params); err != nil {
				l.Error("Error executing list", "error", err)
				exitCode = 1
			}
		}),
	)

	// Set up context with timeout for graceful shutdown
	// This ensures the application doesn't hang indefinitely during shutdown
	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer func() {
		// Attempt to stop the application gracefully
		// Only log non-cancellation errors to avoid noise from normal shutdown
		if err := app.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			if logger != nil {
				logger.Error("Error stopping application", "error", err)
				exitCode = 1
			}
		}
		// Clean up the context after stopping the application
		cancel()
		// Exit with error code if any operation failed
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	// Start the application and handle any startup errors
	if err := app.Start(ctx); err != nil {
		if logger != nil {
			logger.Error("Error starting application", "error", err)
		}
		exitCode = 1
		return
	}
}

// executeList retrieves and displays the list of indices.
// This function handles the core business logic of the list command:
// - Retrieving indices from Elasticsearch
// - Filtering internal indices
// - Formatting and displaying results
//
// Parameters:
//   - p: listParams containing all required dependencies
//
// Returns:
//   - error: Any error encountered during execution
func executeList(p *listParams) error {
	// Retrieve all indices from Elasticsearch
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return err
	}

	// Filter out internal indices (those starting with '.')
	// This improves readability by showing only relevant indices
	var filteredIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	// Handle the case where no indices are found
	if len(filteredIndices) == 0 {
		p.logger.Info("No indices found")
		return nil
	}

	return printIndices(p.ctx, filteredIndices, p.storage, p.logger)
}

// printIndices formats and displays the indices in a table format.
// This function handles the presentation layer of the list command:
// - Creates and configures the output table
// - Retrieves additional metadata for each index
// - Handles errors for individual index operations
// - Formats and displays the final output
//
// Parameters:
//   - ctx: Context for managing timeouts and cancellation
//   - indices: List of index names to display
//   - storage: Interface for Elasticsearch operations
//   - logger: Logger for error reporting
//
// Returns:
//   - error: Any error encountered during table rendering
func printIndices(ctx context.Context, indices []string, storage common.Storage, logger common.Logger) error {
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
		t.AppendRow([]interface{}{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for future ingestion name feature
			ingestionStatus,
		})
	}

	// Handle case where no indices could be processed
	if t.Length() == 0 {
		logger.Info("No indices found")
		return nil
	}

	// Render the final table
	t.Render()
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
