// Package sources implements the command-line interface for managing content sources
// in GoCrawl. It provides commands for listing, adding, removing, and managing
// crawler sources defined in sources.yml.
package sources

import (
	"github.com/spf13/cobra"
)

// Command creates and returns the sources command that serves as the parent
// command for all source management operations. It:
// - Sets up the command with appropriate usage and description
// - Adds subcommands for specific source management operations
// - Provides a unified interface for source configuration management
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage sources defined in sources.yml",
		Long: `Manage sources defined in sources.yml.
These commands help you list, add, remove, and manage your crawler sources.`,
	}

	// Add subcommands for source management operations
	cmd.AddCommand(listCommand())

	return cmd
}
