// Package index implements the command-line interface for managing Elasticsearch
// index in GoCrawl. It provides commands for listing, deleting, and managing
// index in the Elasticsearch cluster.
package index

import (
	"github.com/spf13/cobra"
)

// Command creates and returns the index command that serves as the parent
// command for all index management operations. It:
// - Sets up the command with appropriate usage and description
// - Adds subcommands for specific index management operations
// - Provides a unified interface for index management
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage Elasticsearch index",
		Long: `Manage Elasticsearch index.
This command provides subcommands for listing, deleting, and managing index.`,
	}

	// Add subcommands for index management operations
	cmd.AddCommand(
		NewListCommand(),   // Command for listing all index
		NewDeleteCommand(), // Command for deleting index
		NewCreateCommand(), // Command for creating a new index
	)

	return cmd
}
