package sources

import (
	"github.com/spf13/cobra"
)

// Command creates and returns the sources command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage sources defined in sources.yml",
		Long: `Manage sources defined in sources.yml.
These commands help you list, add, remove, and manage your crawler sources.`,
	}

	// Add subcommands
	cmd.AddCommand(listCommand())

	return cmd
}
