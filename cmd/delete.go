/*
Copyright © 2025 Russell Jones <russell@web.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var deleteSourceName string

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
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

		app := fx.New(
			fx.WithLogger(func() fxevent.Logger {
				return &fxevent.NopLogger
			}),
			fx.Provide(
				func() *config.Config {
					return globalConfig
				},
				func() logger.Interface {
					return globalLogger
				},
			),
			storage.Module,
			sources.Module,
			fx.Invoke(func(storage storage.Interface, sources *sources.Sources) {
				ctx := context.Background()

				// If source is specified, get its indices
				if deleteSourceName != "" {
					source, findErr := sources.FindByName(deleteSourceName)
					if findErr != nil {
						globalLogger.Error("Error finding source", "error", findErr)
						os.Exit(1)
					}
					indices = []string{source.ArticleIndex, source.Index}
				}

				// Check if all indices exist
				existingIndices, listErr := storage.ListIndices(ctx)
				if listErr != nil {
					globalLogger.Error("Error checking indices", "error", listErr)
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
					globalLogger.Info("The following indices do not exist (already deleted):")
					for _, index := range missingIndices {
						globalLogger.Info(fmt.Sprintf("  - %s", index))
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
					globalLogger.Info("Are you sure you want to delete the following indices?")
					for _, index := range indicesToDelete {
						globalLogger.Info(fmt.Sprintf("  - %s", index))
					}
					globalLogger.Info("Continue? (y/N): ")
					var response string
					if _, scanErr := fmt.Scanln(&response); scanErr != nil {
						globalLogger.Error("Error reading response", "error", scanErr)
						return
					}
					if response != "y" && response != "Y" {
						globalLogger.Info("Operation cancelled")
						return
					}
				}

				// Delete each existing index
				for _, index := range indicesToDelete {
					deleteErr := storage.DeleteIndex(ctx, index)
					if deleteErr != nil {
						globalLogger.Error("Error deleting index", "index", index, "error", deleteErr)
						os.Exit(1)
					}
					globalLogger.Info(fmt.Sprintf("Successfully deleted index '%s'", index))
				}
			}),
		)

		if startErr := app.Start(context.Background()); startErr != nil {
			globalLogger.Error("Error starting application", "error", startErr)
			os.Exit(1)
		}

		if stopErr := app.Stop(context.Background()); stopErr != nil {
			globalLogger.Error("Error stopping application", "error", stopErr)
			os.Exit(1)
		}
	},
}

func init() {
	indicesCmd.AddCommand(deleteCmd)

	// Add --force flag to skip confirmation
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	// Add --source flag
	deleteCmd.Flags().StringVar(&deleteSourceName, "source", "", "Delete indices for a specific source")
}
