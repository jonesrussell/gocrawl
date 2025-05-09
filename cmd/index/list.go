// Package index implements the command-line interface for managing Elasticsearch
// index in GoCrawl. This file contains the implementation of the list command
// that displays all index in a formatted table with their health status and metrics.
package index

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
	"github.com/jonesrussell/gocrawl/internal/sources"
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

// RenderTable formats and displays the index in a table format
func (r *TableRenderer) RenderTable(ctx context.Context, storage types.Interface, index []string) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	// Add table headers
	t.AppendHeader(table.Row{"Index", "Health", "Status", "Docs", "Store Size", "Ingestion Status"})

	// Process each index
	for _, index := range index {
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

// Lister handles listing index
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
	// Get all index
	index, err := l.storage.ListIndices(ctx)
	if err != nil {
		l.logger.Error("failed to list index", err)
		return err
	}

	// Filter out internal index (those starting with '.')
	var filteredIndices []string
	for _, index := range index {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	// Render the table
	return l.renderer.RenderTable(ctx, l.storage, filteredIndices)
}

// NewListCommand creates a new list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all index",
		Long:  `List all index in the Elasticsearch cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config from context
			cfgValue := cmd.Context().Value(cmdcommon.ConfigKey)
			cfg, ok := cfgValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context or invalid type")
			}

			// Create Fx application
			app := fx.New(
				// Include required modules
				Module,
				storage.Module,
				sources.Module,

				// Provide existing config
				fx.Provide(func() config.Interface { return cfg }),

				// Provide existing logger
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

// List retrieves and returns a list of all index.
func (l *Lister) List(ctx context.Context) ([]*IndexInfo, error) {
	// Get all index
	index, err := l.storage.ListIndices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list index: %w", err)
	}

	// Filter out internal index (those starting with '.')
	var filteredIndices []string
	for _, index := range index {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	// Get health status and document count for each index
	indexInfo := make([]*IndexInfo, len(filteredIndices))
	for i, index := range filteredIndices {
		// Get health status
		health, healthErr := l.storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			l.logger.Error("Failed to get index health",
				"index", index,
				"error", healthErr,
			)
			continue
		}

		// Get document count
		count, countErr := l.storage.GetIndexDocCount(ctx, index)
		if countErr != nil {
			return nil, fmt.Errorf("failed to get document count for index %s: %w", index, countErr)
		}

		indexInfo[i] = &IndexInfo{
			Name:   index,
			Health: health,
			Count:  count,
		}
	}

	return indexInfo, nil
}

// IndexInfo represents information about an index
type IndexInfo struct {
	Name   string
	Health string
	Count  int64
}
