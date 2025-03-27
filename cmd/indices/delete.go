// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the delete command
// that allows users to delete one or more indices from the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// deleteSourceName holds the name of the source whose indices should be deleted
// when the --source flag is used
var deleteSourceName string

// Params holds the dependencies required for the delete operation.
type Params struct {
	fx.In
	Storage common.Storage
	Sources sources.Interface `name:"sourceManager"`
	Logger  common.Logger
}

// deleteParams holds the parameters required for deleting indices.
type deleteParams struct {
	ctx     context.Context
	storage common.Storage
	sources sources.Interface
	logger  common.Logger
	indices []string
	force   bool
}

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
		common.Module,
		storage.Module,
		fx.Provide(
			func() context.Context { return ctx },
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
					handler.Complete()
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

// executeDelete performs the actual deletion of indices.
// It:
// - Resolves indices to delete (from args or source)
// - Checks which indices exist
// - Filters out non-existent indices
// - Requests confirmation if needed
// - Deletes the specified indices
func executeDelete(p *deleteParams) error {
	if err := resolveIndices(p); err != nil {
		return err
	}

	existingMap, err := getExistingIndices(p)
	if err != nil {
		return err
	}

	indicesToDelete := filterExistingIndices(p, existingMap)
	if len(indicesToDelete) == 0 {
		return nil
	}

	if !p.force {
		if !confirmDeletion(indicesToDelete) {
			return nil
		}
	}

	return deleteIndices(p, indicesToDelete)
}

// resolveIndices determines which indices to delete.
// It:
// - Uses command-line arguments if provided
// - Uses source configuration if --source flag is used
func resolveIndices(p *deleteParams) error {
	if deleteSourceName != "" {
		source, err := p.sources.FindByName(deleteSourceName)
		if err != nil {
			return err
		}
		// Use both content and article indices
		p.indices = []string{source.Index, source.ArticleIndex}
	}
	return nil
}

// getExistingIndices retrieves a list of all existing indices.
// It:
// - Queries Elasticsearch for all indices
// - Creates a map for efficient lookup
func getExistingIndices(p *deleteParams) (map[string]bool, error) {
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, idx := range indices {
		existingMap[idx] = true
	}
	return existingMap, nil
}

// filterExistingIndices filters out non-existent indices and reports them.
// It:
// - Checks each requested index against existing indices
// - Reports non-existent indices to the user
// - Returns only the indices that exist
func filterExistingIndices(p *deleteParams, existingMap map[string]bool) []string {
	var missingIndices []string
	var indicesToDelete []string

	for _, index := range p.indices {
		if !existingMap[index] {
			missingIndices = append(missingIndices, index)
		} else {
			indicesToDelete = append(indicesToDelete, index)
		}
	}

	if len(missingIndices) > 0 {
		common.PrintInfof("\nThe following indices do not exist (already deleted):")
		for _, index := range missingIndices {
			common.PrintInfof("  - %s", index)
		}
	}

	return indicesToDelete
}

// confirmDeletion prompts the user for confirmation before deleting indices.
// It:
// - Displays the list of indices to be deleted
// - Requests user confirmation
func confirmDeletion(indices []string) bool {
	common.PrintInfof("\nAre you sure you want to delete the following indices?")
	for _, index := range indices {
		common.PrintInfof("  - %s", index)
	}
	return common.PrintConfirmation("\nContinue?")
}

// deleteIndices performs the actual deletion of the specified indices.
// It:
// - Deletes each index from Elasticsearch
func deleteIndices(p *deleteParams, indices []string) error {
	for _, index := range indices {
		if err := p.storage.DeleteIndex(p.ctx, index); err != nil {
			return err
		}
	}
	return nil
}
