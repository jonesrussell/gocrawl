package indices

import (
	"github.com/spf13/cobra"
)

// Command returns the indices command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indices",
		Short: "Manage Elasticsearch indices",
		Long: `Manage Elasticsearch indices.
This command provides subcommands for listing, deleting, and managing indices.`,
	}

	// Add subcommands
	cmd.AddCommand(
		listCommand(),
	)

	return cmd
}
