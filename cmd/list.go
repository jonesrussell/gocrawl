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

const (
	// TableWidth is the total width of the table output
	TableWidth = 92
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Elasticsearch indices",
	Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
	Run: func(_ *cobra.Command, _ []string) {
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
					globalLogger.Info("No indices found")
					return
				}

				// Print header
				globalLogger.Info("\nAvailable indices")
				globalLogger.Info("----------------")
				globalLogger.Info(fmt.Sprintf("%-40s %-10s %-12s %-15s %-15s",
					"Index name",
					"Health",
					"Docs count",
					"Ingestion name",
					"Ingestion status"))
				globalLogger.Info(strings.Repeat("-", TableWidth))

				// Print each index
				for _, index := range filteredIndices {
					healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
					if healthErr != nil {
						globalLogger.Error("Error getting health for index", "index", index, "error", healthErr)
						continue
					}

					docCount, docErr := storage.GetIndexDocCount(ctx, index)
					if docErr != nil {
						docCount = 0
					}

					// Determine ingestion status based on health
					ingestionStatus := "Connected"
					if healthStatus == "red" {
						ingestionStatus = "Disconnected"
					} else if healthStatus == "yellow" {
						ingestionStatus = "Warning"
					}

					globalLogger.Info(fmt.Sprintf("%-40s %-10s %-12d %-15s %-15s",
						index,
						healthStatus,
						docCount,
						"", // Ingestion name (not implemented yet)
						ingestionStatus))
				}
				globalLogger.Info("")
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
