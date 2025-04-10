// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the delete command
// that allows users to delete one or more indices from the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	params DeleteParams,
) *Deleter {
	return &Deleter{
		config:     config,
		logger:     logger,
		storage:    storage,
		sources:    sources,
		indices:    params.Indices,
		force:      params.Force,
		sourceName: params.SourceName,
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
			d.logger.Info("Deletion cancelled", "error", err)
			// Return nil for user cancellation to avoid showing usage
			return nil
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

	// Read a single byte to handle EOF and newline cases
	var response [1]byte
	n, err := os.Stdin.Read(response[:])
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	// If no input or newline, treat as 'N'
	if n == 0 || response[0] == '\n' {
		return fmt.Errorf("deletion cancelled by user")
	}

	// Only proceed if input is 'y' or 'Y'
	if response[0] != 'y' && response[0] != 'Y' {
		return fmt.Errorf("deletion cancelled by user")
	}

	return nil
}

// DeleteParams holds the parameters for the delete command
type DeleteParams struct {
	ConfigPath string
	SourceName string
	Force      bool
	Indices    []string
}

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [index-name...]",
		Short: "Delete one or more Elasticsearch indices",
		Long: `Delete one or more Elasticsearch indices.
This command allows you to delete one or more indices from the Elasticsearch cluster.
You can specify indices by name or by source name.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			configPath, _ := cmd.Flags().GetString("config")
			sourceName, _ := cmd.Flags().GetString("source")
			force, _ := cmd.Flags().GetBool("force")

			// Validate args
			if err := ValidateDeleteArgs(sourceName, args); err != nil {
				return err
			}

			// Create Fx application
			app := fx.New(
				// Provide config path string
				fx.Provide(func() string { return configPath }),
				// Provide delete params
				fx.Provide(func() DeleteParams {
					return DeleteParams{
						ConfigPath: configPath,
						SourceName: sourceName,
						Force:      force,
						Indices:    args,
					}
				}),
				// Use the indices module
				Module,
				// Invoke delete command
				fx.Invoke(func(d *Deleter) error {
					return d.Start(cmd.Context())
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
	cmd.Flags().StringP("source", "s", "", "Delete indices for the specified source")
	cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

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
