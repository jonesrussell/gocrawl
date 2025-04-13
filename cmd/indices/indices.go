// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// ConfigKey is the context key for the configuration
	ConfigKey contextKey = "config"
	// LoggerKey is the context key for the logger
	LoggerKey contextKey = "logger"
	// StorageKey is the context key for the storage
	StorageKey contextKey = "storage"
)

// Command creates and returns the indices command that serves as the parent
// command for all index management operations. It:
// - Sets up the command with appropriate usage and description
// - Adds subcommands for specific index management operations
// - Provides a unified interface for index management
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indices",
		Short: "Manage Elasticsearch indices",
		Long: `Manage Elasticsearch indices.
This command provides subcommands for listing, deleting, and managing indices.`,
	}

	// Add subcommands for index management operations
	cmd.AddCommand(
		NewListCommand(),   // Command for listing all indices
		NewDeleteCommand(), // Command for deleting indices
		NewCreateCommand(), // Command for creating a new index
	)

	return cmd
}

// NewIndicesCommand creates a new indices command with dependencies.
func NewIndicesCommand(
	cfg config.Interface,
	log logger.Interface,
	storage types.Interface,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indices",
		Short: "Manage Elasticsearch indices",
		Long: `Manage Elasticsearch indices.
This command provides subcommands for listing, deleting, and managing indices.`,
	}

	// Create a context with dependencies
	ctx := context.Background()
	ctx = context.WithValue(ctx, ConfigKey, cfg)
	ctx = context.WithValue(ctx, LoggerKey, log)
	ctx = context.WithValue(ctx, StorageKey, storage)
	cmd.SetContext(ctx)

	// Add subcommands for index management operations
	cmd.AddCommand(
		NewListCommand(),   // Command for listing all indices
		NewDeleteCommand(), // Command for deleting indices
		NewCreateCommand(), // Command for creating a new index
	)

	return cmd
}
