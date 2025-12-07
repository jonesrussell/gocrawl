// Package index implements the command-line interface for managing Elasticsearch
// index in GoCrawl. This file contains the implementation of the list command
// that displays all index in a formatted table with their health status and metrics.
package index

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/olekukonko/tablewriter"
)

// TableRenderer handles the display of index data in a table format
type TableRenderer struct {
	logger logger.Interface
}

// NewTableRenderer creates a new TableRenderer instance
func NewTableRenderer(log logger.Interface) *TableRenderer {
	return &TableRenderer{
		logger: log,
	}
}

// handleIndexError handles common error cases for index operations
func (r *TableRenderer) handleIndexError(operation, index string, err error, action, details string) error {
	r.logger.Error("Failed to perform index operation",
		"error", err,
		"component", "index",
		"operation", operation,
		"index", index,
		"action", action,
		"details", details,
	)
	return fmt.Errorf("failed to %s for index %s: %w. %s", operation, index, err, action)
}

// RenderTable formats and displays the index in a table format
func (r *TableRenderer) RenderTable(ctx context.Context, storage types.Interface, indices []string) error {
	if len(indices) == 0 {
		r.logger.Info("No indices found in Elasticsearch")
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Health", "Status", "Docs", "Size"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// Add rows
	for _, index := range indices {
		// Get index health
		health, err := storage.GetIndexHealth(ctx, index)
		if err != nil {
			return r.handleIndexError("get health", index, err, "Skipping index", "Failed to retrieve index health")
		}

		// Get document count
		count, err := storage.GetIndexDocCount(ctx, index)
		if err != nil {
			return r.handleIndexError("get doc count", index, err, "Skipping index", "Failed to retrieve document count")
		}

		// Get ingestion status
		ingestionStatus := getIngestionStatus(health)

		// Add row
		table.Append([]string{
			index,
			health,
			ingestionStatus,
			strconv.FormatInt(count, 10),
			"N/A", // Store size not available in current interface
		})
	}

	table.Render()
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
	cfg config.Interface,
	log logger.Interface,
	storage types.Interface,
	renderer *TableRenderer,
) *Lister {
	return &Lister{
		config:   cfg,
		logger:   log,
		storage:  storage,
		renderer: renderer,
	}
}

// Start begins the list operation
func (l *Lister) Start(ctx context.Context) error {
	l.logger.Info("Listing Elasticsearch indices")

	// Get all indices
	indices, err := l.storage.ListIndices(ctx)
	if err != nil {
		l.logger.Error("Failed to list indices",
			"error", err,
			"component", "index",
		)
		return fmt.Errorf("failed to list indices: %w", err)
	}

	// Filter out internal indices (those starting with '.')
	var filteredIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	// Render the table
	return l.renderer.RenderTable(ctx, l.storage, filteredIndices)
}

// getIngestionStatus determines the ingestion status based on health status
func getIngestionStatus(health string) string {
	switch health {
	case "green":
		return "Active"
	case "yellow":
		return "Degraded"
	case "red":
		return "Failed"
	default:
		return "Unknown"
	}
}
