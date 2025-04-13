// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
	t.SetStyle(table.StyleLight)

	// Add table headers
	t.AppendHeader(table.Row{"Index", "Health", "Status", "Docs", "Store Size", "Ingestion Status"})

	// Process each index
	for _, index := range indices {
		// Get index health
		health, err := storage.GetIndexHealth(ctx, index)
		if err != nil {
			return r.handleIndexError("get health", index, err, "Skipping index", "Failed to retrieve index health")
		}

		// Get document count
		docCount, err := storage.GetIndexDocCount(ctx, index)
		if err != nil {
			return r.handleIndexError("get doc count", index, err, "Skipping index", "Failed to retrieve document count")
		}

		// Get ingestion status
		ingestionStatus := getIngestionStatus(health)

		// Add row to table
		t.AppendRow(table.Row{
			index,
			health,
			health,
			docCount,
			"N/A", // Store size not available in current interface
			ingestionStatus,
		})
	}

	// Render the table
	t.Render()
	return nil
}

// Lister handles listing indices
type Lister struct {
	config   config.Interface
	logger   logger.Interface
	storage  types.Interface
	renderer *TableRenderer
}

// NewLister creates a new Lister instance
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

// Start begins the list operation
func (l *Lister) Start(ctx context.Context) error {
	// Get all indices
	indices, err := l.storage.ListIndices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list indices: %w", err)
	}

	// Render the table
	return l.renderer.RenderTable(ctx, l.storage, indices)
}

// NewListCommand creates a new list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all indices",
		Long:  `List all indices in the Elasticsearch cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config path from flags
			configPath, _ := cmd.Flags().GetString("config")

			// Create Fx application
			app := fx.New(
				// Include all required modules
				Module,
				storage.Module,

				// Provide config path string
				fx.Provide(func() string { return configPath }),

				// Provide logger
				fx.Provide(func() logger.Interface { return log }),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke list command
				fx.Invoke(func(l *Lister) error {
					if err := l.Start(cmd.Context()); err != nil {
						return err
					}
					return nil
				}),
			)

			// Start application
			if err := app.Start(context.Background()); err != nil {
				return err
			}

			// Stop application
			if err := app.Stop(context.Background()); err != nil {
				return err
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("config", "c", "config.yaml", "Path to config file")

	return cmd
}

// getIngestionStatus determines the ingestion status based on health status
func getIngestionStatus(healthStatus string) string {
	switch strings.ToLower(healthStatus) {
	case "green":
		return "Active"
	case "yellow":
		return "Degraded"
	case "red":
		return "Stopped"
	default:
		return "Unknown"
	}
}
