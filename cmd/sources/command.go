// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
)

// NewSourcesCommand returns the sources command.
func NewSourcesCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage content sources",
		Long: `The sources command provides functionality for managing content sources.
It allows you to add, list, and configure web content sources for crawling.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newAddCmd(log),
		newListCmd(log),
		newDeleteCmd(log),
	)

	return cmd
}

// newAddCmd creates the add command.
func newAddCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new content source",
		Long: `Add a new content source to the crawler configuration.
The source will be used for future crawling operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Adding new source")
			return nil
		},
	}
}

// newListCmd creates the list command.
func newListCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all content sources",
		Long: `List all configured content sources.
This command shows the details of each source.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Listing sources")
			return nil
		},
	}
}

// newDeleteCmd creates the delete command.
func newDeleteCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a content source",
		Long: `Delete a specific content source by its name.
This command will remove the source from the configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Deleting source")
			return nil
		},
	}
}
