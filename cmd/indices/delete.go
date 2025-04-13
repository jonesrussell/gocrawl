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

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var (
	// ErrDeletionCancelled is returned when the user cancels the deletion
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)

const (
	// defaultIndicesCapacity is the initial capacity for the indices slice
	defaultIndicesCapacity = 2
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
	if err := d.initializeIndices(); err != nil {
		return err
	}

	if err := d.confirmDeletion(); err != nil {
		return err
	}

	return d.deleteIndices(ctx)
}

// initializeIndices initializes the indices slice
func (d *Deleter) initializeIndices() error {
	d.indices = make([]string, 0, defaultIndicesCapacity)
	return nil
}

// confirmDeletion asks for user confirmation before deletion
func (d *Deleter) confirmDeletion() error {
	// Write the list of indices to be deleted
	if _, err := os.Stdout.WriteString("The following indices will be deleted:\n"); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	if _, err := os.Stdout.WriteString(strings.Join(d.indices, "\n") + "\n"); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	if _, err := os.Stdout.WriteString("Are you sure you want to continue? (y/N): "); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	if !strings.EqualFold(response, "y") {
		return ErrDeletionCancelled
	}

	return nil
}

// deleteIndices deletes the indices
func (d *Deleter) deleteIndices(ctx context.Context) error {
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
		if source.Index == "" && source.ArticleIndex == "" {
			return fmt.Errorf("source %s has no indices configured", d.sourceName)
		}
		// Add both content and article indices if they exist
		d.indices = make([]string, 0, defaultIndicesCapacity)
		if source.Index != "" {
			d.indices = append(d.indices, source.Index)
		}
		if source.ArticleIndex != "" {
			d.indices = append(d.indices, source.ArticleIndex)
		}
		d.logger.Info("Resolved source indices", "indices", d.indices, "source", d.sourceName)
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

	// Delete indices
	var deleteErr error
	for _, index := range filtered.toDelete {
		if err := d.storage.DeleteIndex(ctx, index); err != nil {
			d.logger.Error("Failed to delete index",
				"index", index,
				"error", err,
			)
			deleteErr = fmt.Errorf("failed to delete index %s: %w", index, err)
			continue
		}

		d.logger.Info("Successfully deleted index", "index", index)
	}

	if deleteErr != nil {
		return deleteErr
	}

	if len(filtered.toDelete) == 0 {
		d.logger.Info("No indices to delete")
		return nil
	}

	d.logger.Info("Successfully deleted indices", "count", len(filtered.toDelete))
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
		fmt.Fprintf(os.Stdout, "\nThe following indices do not exist:\n")
		for _, index := range missingIndices {
			fmt.Fprintf(os.Stdout, "  - %s\n", index)
		}
	}
}

// DeleteParams holds the parameters for the delete command
type DeleteParams struct {
	ConfigPath string
	SourceName string
	Force      bool
	Indices    []string
}

// NewDeleteCommand creates a new delete command
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [index-name...]",
		Short: "Delete one or more Elasticsearch indices",
		Long: `Delete one or more Elasticsearch indices.
This command allows you to delete one or more indices from the Elasticsearch cluster.
You can specify indices by name or by source name.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

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
				// Include all required modules
				Module,
				storage.Module,

				// Provide config path string
				fx.Provide(func() string { return configPath }),

				// Provide logger
				fx.Provide(func() logger.Interface { return log }),

				// Provide delete params
				fx.Provide(func() DeleteParams {
					return DeleteParams{
						ConfigPath: configPath,
						SourceName: sourceName,
						Force:      force,
						Indices:    args,
					}
				}),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke delete command
				fx.Invoke(func(d *Deleter) error {
					if err := d.Start(cmd.Context()); err != nil {
						if errors.Is(err, ErrDeletionCancelled) {
							cmd.Println("Deletion cancelled")
							return nil
						}
						return err
					}
					cmd.Println("Operation completed successfully")
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
