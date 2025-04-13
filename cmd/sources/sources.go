// Package sources provides the sources command implementation.
package sources

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// ConfigKey is the context key for the configuration
	ConfigKey contextKey = "config"
	// LoggerKey is the context key for the logger
	LoggerKey contextKey = "logger"
)

// NewSourcesCommand returns the sources command.
func NewSourcesCommand(
	cfg config.Interface,
	log logger.Interface,
	_ sources.Interface, // Ignore the passed source manager
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage content sources",
		Long: `The sources command provides functionality for managing content sources.
It allows you to add, list, and configure web content sources for crawling.`,
	}

	// Create a context with dependencies
	ctx := context.Background()
	ctx = context.WithValue(ctx, ConfigKey, cfg)
	ctx = context.WithValue(ctx, LoggerKey, log)
	cmd.SetContext(ctx)

	// Add subcommands
	cmd.AddCommand(
		NewListCommand(),
		// TODO: Implement these commands
		// NewAddCommand(),
		// NewRemoveCommand(),
		// NewUpdateCommand(),
	)

	return cmd
}

// newAddCmd creates the add command.
func newAddCmd(log logger.Interface) *cobra.Command {
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
func newListCmd(log logger.Interface) *cobra.Command {
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
func newDeleteCmd(log logger.Interface) *cobra.Command {
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
