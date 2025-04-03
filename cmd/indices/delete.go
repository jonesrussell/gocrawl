// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the delete command
// that allows users to delete one or more indices from the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// deleteSourceName holds the name of the source whose indices should be deleted
// when the --source flag is used
var deleteSourceName string

// Params holds the dependencies required for the delete operation.
type Params struct {
	fx.In
	Context context.Context `name:"indicesContext"`
	Config  config.Interface
	Storage storagetypes.Interface
	Sources sources.Interface
	Logger  logger.Interface
}

// deleteParams holds the parameters required for deleting indices.
type deleteParams struct {
	ctx     context.Context
	storage storagetypes.Interface
	sources sources.Interface
	logger  logger.Interface
	indices []string
	force   bool
}

// deleteModule provides the delete command dependencies
var deleteModule = fx.Module("delete",
	// Core dependencies
	config.Module,
	storage.Module,
	sources.Module,
)

// deleteCommand creates and returns the delete command that removes indices.
// It:
// - Sets up the command with appropriate usage and description
// - Adds command-line flags for source and force options
// - Configures argument validation
// - Configures the command to use runDelete as its execution function
func deleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [indices...]",
		Short: "Delete one or more Elasticsearch indices",
		Long: `Delete one or more Elasticsearch indices from the cluster.
If --source is specified, deletes the indices associated with that source.

Example:
  gocrawl indices delete my_index
  gocrawl indices delete index1 index2 index3
  gocrawl indices delete --source "Elliot Lake Today"`,
		Args: validateDeleteArgs,
		RunE: runDelete,
	}

	// Add command-line flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&deleteSourceName, "source", "", "Delete indices for a specific source")

	return cmd
}

// validateDeleteArgs validates the command arguments to ensure they are valid.
// It:
// - Ensures either indices are specified or --source flag is used
// - Prevents using both indices and --source flag together
func validateDeleteArgs(_ *cobra.Command, args []string) error {
	if deleteSourceName == "" && len(args) == 0 {
		return errors.New("either specify indices or use --source flag")
	}
	if deleteSourceName != "" && len(args) > 0 {
		return errors.New("cannot specify both indices and --source flag")
	}
	// Trim quotes from source name if present
	if deleteSourceName != "" {
		deleteSourceName = strings.Trim(deleteSourceName, "\"")
	}
	return nil
}

// runDelete executes the delete command and removes the specified indices.
func runDelete(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	// Create a cancellable context
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Set up signal handling with a no-op logger initially
	handler := signal.NewSignalHandler(logger.NewNoOp())
	handler.Setup(ctx)

	// Initialize the Fx application
	fxApp := fx.New(
		fx.NopLogger,
		deleteModule,
		fx.Provide(
			func() context.Context { return ctx },
			logger.NewNoOp,
		),
		fx.Invoke(func(lc fx.Lifecycle, p Params) {
			// Update the signal handler with the real logger
			handler.SetLogger(p.Logger)
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &deleteParams{
						ctx:     ctx,
						storage: p.Storage,
						sources: p.Sources,
						logger:  p.Logger,
						indices: args,
						force:   force,
					}
					if err := executeDelete(params); err != nil {
						p.Logger.Error("Error executing delete", "error", err)
						return err
					}
					// Signal completion
					handler.RequestShutdown()
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Set the fx app for coordinated shutdown
	handler.SetFXApp(fxApp)

	// Start the application
	if err := fxApp.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Wait for shutdown signal
	handler.Wait()

	return nil
}

// filterIndices filters out non-existent indices and returns lists of indices to delete and missing indices.
func filterIndices(p *deleteParams, existingIndices []string) (indicesToDelete, missingIndices []string) {
	// Create map of existing indices
	existingMap := make(map[string]bool)
	for _, idx := range existingIndices {
		existingMap[idx] = true
	}

	// Filter and report non-existent indices
	for _, index := range p.indices {
		if !existingMap[index] {
			missingIndices = append(missingIndices, index)
		} else {
			indicesToDelete = append(indicesToDelete, index)
		}
	}

	return indicesToDelete, missingIndices
}

// reportMissingIndices prints a list of indices that do not exist.
func reportMissingIndices(missingIndices []string) {
	if len(missingIndices) > 0 {
		fmt.Fprintf(os.Stdout, "\nThe following indices do not exist (already deleted):\n")
		for _, index := range missingIndices {
			fmt.Fprintf(os.Stdout, "  - %s\n", index)
		}
	}
}

// confirmDeletion prompts the user to confirm deletion of indices.
func confirmDeletion(indicesToDelete []string) error {
	fmt.Fprintf(os.Stdout, "\nAre you sure you want to delete the following indices?\n")
	for _, index := range indicesToDelete {
		fmt.Fprintf(os.Stdout, "  - %s\n", index)
	}
	fmt.Fprintf(os.Stdout, "\nContinue? (y/N): ")
	var response string
	if _, confirmErr := fmt.Scanln(&response); confirmErr != nil {
		return fmt.Errorf("failed to read user input: %w", confirmErr)
	}
	if response != "y" && response != "Y" {
		return nil
	}
	return nil
}

// executeDelete performs the actual deletion of indices.
func executeDelete(p *deleteParams) error {
	// Resolve indices to delete
	if deleteSourceName != "" {
		source, err := p.sources.FindByName(deleteSourceName)
		if err != nil {
			return err
		}
		p.indices = []string{source.Index, source.ArticleIndex}
	}

	// Get existing indices
	existingIndices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return err
	}

	// Filter indices
	indicesToDelete, missingIndices := filterIndices(p, existingIndices)

	// Report missing indices
	reportMissingIndices(missingIndices)

	if len(indicesToDelete) == 0 {
		return nil
	}

	// Confirm deletion if needed
	if !p.force {
		if confirmErr := confirmDeletion(indicesToDelete); confirmErr != nil {
			return confirmErr
		}
	}

	// Delete indices
	for _, index := range indicesToDelete {
		if deleteErr := p.storage.DeleteIndex(p.ctx, index); deleteErr != nil {
			return deleteErr
		}
	}
	return nil
}
