// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the delete command
// that allows users to delete one or more indices from the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Deleter implements the indices delete command
type Deleter struct {
	config     config.Interface
	logger     logger.Interface
	storage    storagetypes.Interface
	sources    sources.Interface
	indices    []string
	force      bool
	sourceName string
}

// NewDeleter creates a new deleter instance
func NewDeleter(
	config config.Interface,
	logger logger.Interface,
	storage storagetypes.Interface,
	sources sources.Interface,
	indices []string,
	force bool,
	sourceName string,
) *Deleter {
	return &Deleter{
		config:     config,
		logger:     logger,
		storage:    storage,
		sources:    sources,
		indices:    indices,
		force:      force,
		sourceName: sourceName,
	}
}

// Start executes the delete operation
func (d *Deleter) Start(ctx context.Context) error {
	d.logger.Info("Starting index deletion", "indices", d.indices, "source", d.sourceName)

	// Test storage connection
	if err := d.storage.TestConnection(ctx); err != nil {
		d.logger.Error("Failed to connect to storage", "error", err)
		return fmt.Errorf("failed to connect to storage: %w", err)
	}

	// Resolve indices to delete
	if d.sourceName != "" {
		source := d.sources.FindByName(d.sourceName)
		if source == nil {
			return fmt.Errorf("source not found: %s", d.sourceName)
		}
		d.indices = []string{source.Index}
		d.logger.Info("Resolved source indices", "indices", d.indices)
	}

	// Get existing indices
	existingIndices, listErr := d.storage.ListIndices(ctx)
	if listErr != nil {
		d.logger.Error("Failed to list indices", "error", listErr)
		return listErr
	}
	d.logger.Debug("Found existing indices", "indices", existingIndices)

	// Check for empty indices
	if len(d.indices) == 0 {
		return errors.New("no indices specified")
	}

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
		if err := d.storage.DeleteIndex(ctx, index); err != nil {
			d.logger.Error("Failed to delete index", "index", index, "error", err)
			return fmt.Errorf("failed to delete index %s: %w", index, err)
		}
		d.logger.Info("Successfully deleted index", "index", index)
	}

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

// deleteCommand creates and returns the command for deleting Elasticsearch indices.
func deleteCommand() *cobra.Command {
	var (
		force      bool
		sourceName string
	)

	cmd := &cobra.Command{
		Use:   "delete [indices...]",
		Short: "Delete one or more Elasticsearch indices",
		Long: `Delete one or more Elasticsearch indices.
		
You can specify one or more indices to delete, or use the --source flag to delete indices for a specific source.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ValidateDeleteArgs(sourceName, args); err != nil {
				return err
			}

			app := fx.New(
				fx.NopLogger,
				Module,
				fx.Provide(
					func() []string { return args },
					func() bool { return force },
					func() string { return sourceName },
					NewDeleter,
				),
				fx.Invoke(func(deleter *Deleter) error {
					return deleter.Start(cmd.Context())
				}),
			)

			return app.Start(cmd.Context())
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVarP(&sourceName, "source", "s", "", "Delete indices for a specific source")

	return cmd
}

// ValidateDeleteArgs validates the arguments for the delete command.
func ValidateDeleteArgs(sourceName string, args []string) error {
	if sourceName == "" && len(args) == 0 {
		return errors.New("either indices or a source name must be specified")
	}
	if sourceName != "" && len(args) > 0 {
		return errors.New("cannot specify both indices and a source name")
	}
	return nil
}
