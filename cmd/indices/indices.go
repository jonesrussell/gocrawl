// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"github.com/spf13/cobra"
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
		listCommand(),   // Command for listing all indices
		deleteCommand(), // Command for deleting indices
		createCommand(), // Command for creating a new index
	)

	return cmd
}
