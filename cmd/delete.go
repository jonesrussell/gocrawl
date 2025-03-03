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
	Use:   "delete [index]",
	Short: "Delete an Elasticsearch index",
	Long: `Delete an Elasticsearch index from the cluster.

Example:
  gocrawl indices delete my_index`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		index := args[0]
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

				// Check if index exists
				indices, err := storage.ListIndices(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error checking indices: %v\n", err)
					os.Exit(1)
				}

				indexExists := false
				for _, idx := range indices {
					if idx == index {
						indexExists = true
						break
					}
				}

				if !indexExists {
					fmt.Fprintf(os.Stderr, "Index '%s' does not exist\n", index)
					os.Exit(1)
				}

				// Confirm deletion unless --force is used
				if !force {
					fmt.Printf("Are you sure you want to delete index '%s'? (y/N): ", index)
					var response string
					fmt.Scanln(&response)
					if response != "y" && response != "Y" {
						fmt.Println("Operation cancelled")
						return
					}
				}

				// Delete the index
				if err := storage.DeleteIndex(ctx, index); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting index: %v\n", err)
					os.Exit(1)
				}

				fmt.Printf("Successfully deleted index '%s'\n", index)
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
