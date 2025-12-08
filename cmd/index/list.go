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

	"github.com/jedib0t/go-pretty/v6/table"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
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
func (r *TableRenderer) RenderTable(ctx context.Context, stor types.Interface, indices []string) error {
	if len(indices) == 0 {
		r.logger.Info("No indices found in Elasticsearch")
		return nil
	}

	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Configure table to match original tablewriter behavior:
	// - No borders or separators (plain text format)
	// - Tab padding between columns
	noBorderStyle := table.Style{
		Box: table.BoxStyle{
			BottomLeft:       "",
			BottomRight:      "",
			BottomSeparator:  "",
			Left:             "",
			LeftSeparator:    "",
			MiddleHorizontal: "",
			MiddleSeparator:  "",
			MiddleVertical:   "",
			PaddingLeft:      "\t",
			PaddingRight:     "\t",
			Right:            "",
			RightSeparator:   "",
			TopLeft:          "",
			TopRight:         "",
			TopSeparator:     "",
			UnfinishedRow:    "",
		},
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: false,
			SeparateHeader:  false,
			SeparateRows:    false,
		},
	}
	t.SetStyle(noBorderStyle)

	// Add table headers
	t.AppendHeader(table.Row{"Index", "Health", "Status", "Docs", "Size"})

	// Add rows
	for _, index := range indices {
		// Get index health
		health, err := stor.GetIndexHealth(ctx, index)
		if err != nil {
			return r.handleIndexError("get health", index, err, "Skipping index", "Failed to retrieve index health")
		}

		// Get document count
		count, err := stor.GetIndexDocCount(ctx, index)
		if err != nil {
			return r.handleIndexError("get doc count", index, err, "Skipping index", "Failed to retrieve document count")
		}

		// Get ingestion status
		ingestionStatus := getIngestionStatus(health)

		// Add row
		t.AppendRow(table.Row{
			index,
			health,
			ingestionStatus,
			strconv.FormatInt(count, 10),
			"N/A", // Store size not available in current interface
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
	cfg config.Interface,
	log logger.Interface,
	stor types.Interface,
	renderer *TableRenderer,
) *Lister {
	return &Lister{
		config:   cfg,
		logger:   log,
		storage:  stor,
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

// runListCmd executes the list command
func runListCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get dependencies
	deps, err := cmdcommon.NewCommandDeps()
	if err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Create storage using common function
	storageResult, err := cmdcommon.CreateStorage(deps.Config, deps.Logger)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	renderer := NewTableRenderer(deps.Logger)
	lister := NewLister(deps.Config, deps.Logger, storageResult.Storage, renderer)

	return lister.Start(ctx)
}
