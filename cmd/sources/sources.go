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
