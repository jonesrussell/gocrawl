// Package index implements the command-line interface for managing Elasticsearch
// index in GoCrawl. It provides commands for listing, deleting, and managing
// index in the Elasticsearch cluster.
package index

import (
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
	sourceName  string
)

// Command returns the index command for use in the root command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage Elasticsearch indices",
		Long:  `Manage Elasticsearch indices for storing crawled content`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(createListCmd(), createCreateCmd(), createDeleteCmd())
	return cmd
}

// createListCmd creates the list command
func createListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all indices",
		RunE:  runListCmd,
	}
}

// createCreateCmd creates the create command
func createCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [index-name]",
		Short: "Create an index",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreateCmd,
	}
}

// createDeleteCmd creates the delete command
func createDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [index-name...]",
		Short: "Delete an index",
		Long: `Delete one or more indices. Either provide index names as arguments ` +
			`or use --source to delete indices for a specific source.`,
		Args: cobra.MinimumNArgs(0),
		RunE: runDeleteCmd,
	}
	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion without confirmation")
	cmd.Flags().StringVar(&sourceName, "source", "", "Delete index for a specific source by name")
	return cmd
}
