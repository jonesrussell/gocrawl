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
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Elasticsearch indices",
	Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
	Run: func(cmd *cobra.Command, args []string) {
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
				indices, err := storage.ListIndices(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error listing indices: %v\n", err)
					os.Exit(1)
				}

				// Filter out internal indices
				var filteredIndices []string
				for _, index := range indices {
					if !strings.HasPrefix(index, ".") {
						filteredIndices = append(filteredIndices, index)
					}
				}

				if len(filteredIndices) == 0 {
					fmt.Println("No indices found")
					return
				}

				// Print header
				fmt.Println("\nAvailable indices")
				fmt.Println("----------------")
				fmt.Printf("%-40s %-10s %-12s %-15s %-15s\n",
					"Index name",
					"Health",
					"Docs count",
					"Ingestion name",
					"Ingestion status")
				fmt.Println(strings.Repeat("-", 92))

				// Print each index
				for _, index := range filteredIndices {
					health, err := storage.GetIndexHealth(ctx, index)
					if err != nil {
						health = "unknown"
					}

					docCount, err := storage.GetIndexDocCount(ctx, index)
					if err != nil {
						docCount = 0
					}

					// Determine ingestion status based on health
					ingestionStatus := "Connected"
					if health == "red" {
						ingestionStatus = "Disconnected"
					} else if health == "yellow" {
						ingestionStatus = "Warning"
					}

					fmt.Printf("%-40s %-10s %-12d %-15s %-15s\n",
						index,
						health,
						docCount,
						"", // Ingestion name (not implemented yet)
						ingestionStatus)
				}
				fmt.Println()
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
	indicesCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
