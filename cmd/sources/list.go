// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"fmt"

	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params holds the dependencies required for the list operation.
type Params struct {
	fx.In
	SourceManager sources.Interface
	Logger        logger.Interface
}

// ListCommand implements the list command for sources.
type ListCommand struct {
	logger        logger.Interface
	sourceManager sources.Interface
}

// NewListCommand creates a new list command.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long:  `List all content sources configured in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get dependencies from context
			loggerValue := cmd.Context().Value(LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context")
			}

			configValue := cmd.Context().Value(ConfigKey)
			cfg, ok := configValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context")
			}

			// Create source manager
			sourceManager, err := sources.LoadSources(cfg)
			if err != nil {
				return fmt.Errorf("failed to load sources: %w", err)
			}

			// Create and run the command
			listCmd := &ListCommand{
				logger:        log,
				sourceManager: sourceManager,
			}
			return listCmd.Run(cmd.Context())
		},
	}

	return cmd
}

// Run executes the list command.
func (c *ListCommand) Run(ctx context.Context) error {
	c.logger.Info("Listing sources")

	// Get all sources
	sourceList, err := c.sourceManager.ListSources(ctx)
	if err != nil {
		c.logger.Error("Failed to list sources", "error", err)
		return fmt.Errorf("failed to list sources: %w", err)
	}

	// Print sources
	if len(sourceList) == 0 {
		c.logger.Info("No sources found")
		return nil
	}

	for _, source := range sourceList {
		c.logger.Info("Source found",
			"name", source.Name,
			"url", source.URL,
			"allowed_domains", source.AllowedDomains,
			"start_urls", source.StartURLs,
			"max_depth", source.MaxDepth,
			"rate_limit", source.RateLimit,
			"index", source.Index,
			"article_index", source.ArticleIndex,
		)
	}

	return nil
}

// RunList executes the list command and displays all sources.
func RunList(cmd *cobra.Command, _ []string) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Set up signal handling with a no-op logger initially
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules
	app := fx.New(
		fx.NopLogger,
		fx.Supply(cmd),
		fx.Invoke(func(p struct {
			fx.In
			Sources sources.Interface
			Logger  logger.Interface
			LC      fx.Lifecycle
		}) {
			// Update the signal handler with the real logger
			handler.SetLogger(p.Logger)
			p.LC.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &Params{
						SourceManager: p.Sources,
						Logger:        p.Logger,
					}
					if err := ExecuteList(*params, ctx); err != nil {
						p.Logger.Error("Error executing list", "error", err)
						errChan <- err
						return err
					}
					close(doneChan)
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the application and handle any startup errors
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Set up cleanup for graceful shutdown
	handler.SetCleanup(func() {
		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultOperationTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
		}
	})

	// Wait for either:
	// - A signal interrupt (SIGINT/SIGTERM)
	// - Context cancellation
	// - List completion
	// - List error
	var listErr error
	select {
	case listErr = <-errChan:
		// Error already logged in ExecuteList
	case <-doneChan:
		// Success message already printed in ExecuteList
	}

	// Only wait for shutdown signal if there was an error
	if listErr != nil {
		if err := handler.Wait(); err != nil {
			return fmt.Errorf("failed to wait for handler: %w", err)
		}
	}

	return listErr
}

// ExecuteList retrieves and displays the list of sources.
func ExecuteList(params Params, ctx context.Context) error {
	// Get all sources
	allSources, err := params.SourceManager.ListSources(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	// Convert []*Config to []Config
	sources := make([]sources.Config, len(allSources))
	for i, src := range allSources {
		sources[i] = *src
	}

	// Print sources
	if printErr := PrintSources(sources, params.Logger); printErr != nil {
		return fmt.Errorf("failed to print sources: %w", printErr)
	}
	return nil
}

// PrintSources prints the list of sources.
func PrintSources(sources []sources.Config, logger logger.Interface) error {
	if len(sources) == 0 {
		logger.Info("No sources found")
		return nil
	}

	logger.Info("Found sources", "count", len(sources))
	for i := range sources {
		logger.Info("Source", "name", sources[i].Name, "url", sources[i].URL)
	}

	return nil
}
