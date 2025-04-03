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

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// deleteSourceName holds the name of the source whose indices should be deleted
// when the --source flag is used
var deleteSourceName string

// Deleter implements the indices delete command
type Deleter struct {
	config  config.Interface
	logger  logger.Interface
	storage storagetypes.Interface
	sources sources.Interface
	indices []string
	force   bool
}

// NewDeleter creates a new deleter instance
func NewDeleter(
	config config.Interface,
	logger logger.Interface,
	storage storagetypes.Interface,
	sources sources.Interface,
	indices []string,
	force bool,
) *Deleter {
	return &Deleter{
		config:  config,
		logger:  logger,
		storage: storage,
		sources: sources,
		indices: indices,
		force:   force,
	}
}

// Start executes the delete operation
func (d *Deleter) Start(ctx context.Context) error {
	d.logger.Info("Starting index deletion", "indices", d.indices, "source", deleteSourceName)

	// Resolve indices to delete
	if deleteSourceName != "" {
		source := d.sources.FindByName(deleteSourceName)
		if source == nil {
			return fmt.Errorf("source not found: %s", deleteSourceName)
		}
		d.indices = []string{source.Index, source.ArticleIndex}
		d.logger.Info("Resolved source indices", "indices", d.indices)
	}

	// Get existing indices
	existingIndices, err := d.storage.ListIndices(ctx)
	if err != nil {
		d.logger.Error("Failed to list indices", "error", err)
		return err
	}
	d.logger.Debug("Found existing indices", "indices", existingIndices)

	// Filter indices
	filtered := d.filterIndices(existingIndices)

	// Report missing indices
	d.reportMissingIndices(filtered.missing)

	if len(filtered.toDelete) == 0 {
		d.logger.Info("No indices to delete")
		return nil
	}

	d.logger.Info("Indices to delete", "indices", filtered.toDelete)

	// Confirm deletion if needed
	if !d.force {
		if err := d.confirmDeletion(filtered.toDelete); err != nil {
			return err
		}
	}

	// Delete indices
	for _, index := range filtered.toDelete {
		d.logger.Info("Deleting index", "index", index)
		if err := d.storage.DeleteIndex(ctx, index); err != nil {
			d.logger.Error("Failed to delete index", "index", index, "error", err)
			return err
		}
		d.logger.Info("Successfully deleted index", "index", index)
	}

	d.logger.Info("Index deletion completed successfully")
	return nil
}

// filterIndices filters out non-existent indices and returns lists of indices to delete and missing indices.
func (d *Deleter) filterIndices(existingIndices []string) struct {
	toDelete []string
	missing  []string
} {
	// Create map of existing indices
	existingMap := make(map[string]bool)
	for _, idx := range existingIndices {
		existingMap[idx] = true
	}

	// Filter and report non-existent indices
	result := struct {
		toDelete []string
		missing  []string
	}{
		toDelete: make([]string, 0, len(d.indices)),
		missing:  make([]string, 0, len(d.indices)),
	}

	for _, index := range d.indices {
		if !existingMap[index] {
			result.missing = append(result.missing, index)
		} else {
			result.toDelete = append(result.toDelete, index)
		}
	}

	return result
}

// reportMissingIndices prints a list of indices that do not exist.
func (d *Deleter) reportMissingIndices(missingIndices []string) {
	if len(missingIndices) > 0 {
		fmt.Fprintf(os.Stdout, "\nThe following indices do not exist (already deleted):\n")
		for _, index := range missingIndices {
			fmt.Fprintf(os.Stdout, "  - %s\n", index)
		}
	}
}

// confirmDeletion prompts the user to confirm deletion of indices.
func (d *Deleter) confirmDeletion(indicesToDelete []string) error {
	fmt.Fprintf(os.Stdout, "\nAre you sure you want to delete the following indices?\n")
	for _, index := range indicesToDelete {
		fmt.Fprintf(os.Stdout, "  - %s\n", index)
	}
	fmt.Fprintf(os.Stdout, "\nContinue? (y/N): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}
	if response != "y" && response != "Y" {
		return nil
	}
	return nil
}

// deleteCommand creates and returns the delete command that removes indices.
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

	// Initialize the Fx application
	fxApp := fx.New(
		fx.NopLogger,
		Module,
		fx.Provide(
			func() context.Context { return ctx },
			func() []string { return args },
			func() bool { return force },
		),
		fx.Invoke(func(lc fx.Lifecycle, deleter *Deleter) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := deleter.Start(ctx); err != nil {
						deleter.logger.Error("Error executing delete", "error", err)
						return err
					}
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the application
	if err := fxApp.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Stop the application
	if err := fxApp.Stop(ctx); err != nil {
		return fmt.Errorf("error stopping application: %w", err)
	}

	return nil
}
