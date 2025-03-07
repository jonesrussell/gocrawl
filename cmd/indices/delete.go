package indices

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var deleteSourceName string

// deleteCommand returns the delete command
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
		Args: func(_ *cobra.Command, args []string) error {
			if deleteSourceName == "" && len(args) == 0 {
				return errors.New("either specify indices or use --source flag")
			}
			if deleteSourceName != "" && len(args) > 0 {
				return errors.New("cannot specify both indices and --source flag")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			indices := args
			force, _ := cmd.Flags().GetBool("force")

			var logger common.Logger // Capture logger for lifecycle errors

			app := fx.New(
				common.Module,
				fx.Invoke(func(storage common.Storage, sources common.Sources, l common.Logger) {
					logger = l
					ctx := context.Background()

					// If source is specified, get its indices
					if deleteSourceName != "" {
						source, findErr := sources.FindByName(deleteSourceName)
						if findErr != nil {
							l.Error("Error finding source", "error", findErr)
							os.Exit(1)
						}
						indices = []string{source.ArticleIndex, source.Index}
					}

					// Check if all indices exist
					existingIndices, listErr := storage.ListIndices(ctx)
					if listErr != nil {
						l.Error("Error checking indices", "error", listErr)
						os.Exit(1)
					}

					// Create a map of existing indices for faster lookup
					existingMap := make(map[string]bool)
					for _, idx := range existingIndices {
						existingMap[idx] = true
					}

					// Check each requested index
					var missingIndices []string
					for _, index := range indices {
						if !existingMap[index] {
							missingIndices = append(missingIndices, index)
						}
					}

					if len(missingIndices) > 0 {
						l.Info("The following indices do not exist (already deleted):")
						for _, index := range missingIndices {
							l.Info(fmt.Sprintf("  - %s", index))
						}
						if len(missingIndices) == len(indices) {
							return // All indices are already deleted, exit successfully
						}
					}

					// Get list of indices that do exist and need to be deleted
					var indicesToDelete []string
					for _, index := range indices {
						if existingMap[index] {
							indicesToDelete = append(indicesToDelete, index)
						}
					}

					// Confirm deletion unless --force is used
					if !force && len(indicesToDelete) > 0 {
						l.Info("Are you sure you want to delete the following indices?")
						for _, index := range indicesToDelete {
							l.Info(fmt.Sprintf("  - %s", index))
						}
						l.Info("Continue? (y/N): ")
						var response string
						if _, scanErr := fmt.Scanln(&response); scanErr != nil {
							l.Error("Error reading response", "error", scanErr)
							return
						}
						if response != "y" && response != "Y" {
							l.Info("Operation cancelled")
							return
						}
					}

					// Delete each existing index
					for _, index := range indicesToDelete {
						deleteErr := storage.DeleteIndex(ctx, index)
						if deleteErr != nil {
							l.Error("Error deleting index", "index", index, "error", deleteErr)
							os.Exit(1)
						}
						l.Info(fmt.Sprintf("Successfully deleted index '%s'", index))
					}
				}),
			)

			// Use Cobra's context with timeout for lifecycle handling
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			if startErr := app.Start(ctx); startErr != nil {
				if logger != nil {
					logger.Error("Error starting application", "error", startErr)
				}
				os.Exit(1)
			}

			if stopErr := app.Stop(ctx); stopErr != nil {
				if logger != nil {
					logger.Error("Error stopping application", "error", stopErr)
				}
				os.Exit(1)
			}
		},
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&deleteSourceName, "source", "", "Delete indices for a specific source")

	return cmd
}
