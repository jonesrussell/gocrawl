// Package sources provides the sources command implementation.
package sources

import (
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common"
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

// NewSourcesCommand creates a new sources command
func NewSourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage content sources",
		Long:  `Manage content sources for crawling`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			sourceManager, err := sources.LoadSources(cfg)
			if err != nil {
				return fmt.Errorf("failed to load sources: %w", err)
			}

			// List all sources
			srcs, err := sourceManager.ListSources(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sources: %w", err)
			}

			if len(srcs) == 0 {
				loggerValue, ok := ctx.Value(common.LoggerKey).(logger.Interface)
				if !ok {
					return errors.New("logger not found in context or invalid type")
				}
				loggerValue.Info("No sources configured")
				return nil
			}

			loggerValue, ok := ctx.Value(common.LoggerKey).(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}
			loggerValue.Info("Configured sources:")
			for _, src := range srcs {
				loggerValue.Info("- " + src.Name)
			}

			return nil
		},
	}

	return cmd
}
