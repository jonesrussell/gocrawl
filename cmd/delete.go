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
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [indices...]",
	Short: "Delete one or more Elasticsearch indices",
	Long: `Delete one or more Elasticsearch indices from the cluster.

Example:
  gocrawl indices delete my_index
  gocrawl indices delete index1 index2 index3`,
	Args: cobra.MinimumNArgs(1),
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
			fx.Invoke(func(storage storage.Interface) {
				ctx := context.Background()

				// Check if all indices exist
				existingIndices, err := storage.ListIndices(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error checking indices: %v\n", err)
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
					fmt.Fprintf(os.Stderr, "The following indices do not exist:\n")
					for _, index := range missingIndices {
						fmt.Fprintf(os.Stderr, "  - %s\n", index)
					}
					os.Exit(1)
				}

				// Confirm deletion unless --force is used
				if !force {
					fmt.Printf("Are you sure you want to delete the following indices?\n")
					for _, index := range indices {
						fmt.Printf("  - %s\n", index)
					}
					fmt.Print("Continue? (y/N): ")
					var response string
					fmt.Scanln(&response)
					if response != "y" && response != "Y" {
						fmt.Println("Operation cancelled")
						return
					}
				}

				// Delete each index
				for _, index := range indices {
					if err := storage.DeleteIndex(ctx, index); err != nil {
						fmt.Fprintf(os.Stderr, "Error deleting index '%s': %v\n", index, err)
						os.Exit(1)
					}
					fmt.Printf("Successfully deleted index '%s'\n", index)
				}
			}),
		)

		if err := app.Start(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting application: %v\n", err)
			os.Exit(1)
		}

		if err := app.Stop(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping application: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	indicesCmd.AddCommand(deleteCmd)

	// Add --force flag to skip confirmation
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
