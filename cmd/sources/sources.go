// Package sources provides the sources command implementation.
package sources

import (
	"context"
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// SourcesCommand implements the sources command.
type SourcesCommand struct {
	sourceManager sources.Interface
	logger        logger.Interface
}

// NewSourcesCommand creates a new sources command.
func NewSourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage content sources",
		Long:  `Manage content sources for crawling`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config from context
			configValue := cmd.Context().Value(cmdcommon.ConfigKey)
			cfg, ok := configValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context or invalid type")
			}

			// Create Fx app with the module
			fxApp := fx.New(
				// Include required modules
				sources.Module,

				// Provide existing config
				fx.Provide(func() config.Interface { return cfg }),

				// Provide existing logger
				fx.Provide(func() logger.Interface { return log }),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke sources command
				fx.Invoke(func(sourceManager sources.Interface, logger logger.Interface) error {
					sourcesCmd := &SourcesCommand{
						sourceManager: sourceManager,
						logger:        logger,
					}
					return sourcesCmd.Run(cmd.Context())
				}),
			)

			// Start the application
			log.Info("Starting application")
			startErr := fxApp.Start(cmd.Context())
			if startErr != nil {
				log.Error("Failed to start application", "error", startErr)
				return fmt.Errorf("failed to start application: %w", startErr)
			}

			// Stop the application
			log.Info("Stopping application")
			stopErr := fxApp.Stop(cmd.Context())
			if stopErr != nil {
				log.Error("Failed to stop application", "error", stopErr)
				return fmt.Errorf("failed to stop application: %w", stopErr)
			}

			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(
		NewListCommand(),
	)

	return cmd
}

// Run executes the sources command.
func (c *SourcesCommand) Run(ctx context.Context) error {
	c.logger.Info("Listing sources")

	sources, err := c.sourceManager.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sources) == 0 {
		c.logger.Info("No sources configured")
		return nil
	}

	// Print sources in a formatted table
	log := c.logger
	log.Info("Configured Sources:")
	log.Info("------------------")
	for _, src := range sources {
		log.Info("Source details",
			"name", src.Name,
			"url", src.URL,
			"allowed_domains", src.AllowedDomains,
			"start_urls", src.StartURLs,
			"max_depth", src.MaxDepth,
			"rate_limit", src.RateLimit,
		)
		log.Info("------------------")
	}

	return nil
}
