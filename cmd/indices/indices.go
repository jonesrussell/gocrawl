// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(LoggerKey)
			loggerInterface, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config path from flags
			configPath, _ := cmd.Flags().GetString("config")

			// Create Fx application
			app := fx.New(
				// Include all required modules
				Module,

				// Provide config path string
				fx.Provide(func() string { return configPath }),

				// Provide logger
				fx.Provide(func() logger.Interface { return loggerInterface }),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(loggerInterface)
				}),

				// Invoke indices command
				fx.Invoke(func(cmd *cobra.Command) error {
					// Add subcommands for index management operations
					cmd.AddCommand(
						NewListCommand(),   // Command for listing all indices
						NewDeleteCommand(), // Command for deleting indices
						NewCreateCommand(), // Command for creating a new index
					)
					return nil
				}),
			)

			// Start application
			if err := app.Start(context.Background()); err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			// Stop application
			if err := app.Stop(context.Background()); err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("config", "c", "config.yaml", "Path to config file")

	return cmd
}
