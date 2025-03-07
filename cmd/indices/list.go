package indices

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// TableWidth is the total width of the table output
	TableWidth = 92
)

// listCommand returns the list command
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		Run: func(cmd *cobra.Command, _ []string) {
			var logger common.Logger // Capture logger for lifecycle errors

			app := fx.New(
				common.Module,
				fx.Invoke(func(s common.Storage, l common.Logger) {
					logger = l // Store logger for lifecycle errors

					ctx := cmd.Context() // Use Cobra's context
					indices, err := s.ListIndices(ctx)
					if err != nil {
						l.Error("Error listing indices", "error", err)
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
						l.Info("No indices found")
						return
					}

					printIndices(filteredIndices, s, l, ctx) // Use helper function
				}),
			)

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second) // Combine Cobra's context with a timeout
			defer cancel()

			if err := app.Start(ctx); err != nil {
				if logger != nil {
					logger.Error("Error starting application", "error", err)
				}
				os.Exit(1)
			}

			if err := app.Stop(ctx); err != nil {
				if logger != nil {
					logger.Error("Error stopping application", "error", err)
				}
				os.Exit(1)
			}
		},
	}
}

func printIndices(indices []string, storage common.Storage, logger common.Logger, ctx context.Context) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Ingestion Name", "Ingestion Status"})

	for _, index := range indices {
		healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			logger.Error("Error getting health for index", "index", index, "error", healthErr)
			continue
		}

		docCount, docErr := storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			docCount = 0
		}

		ingestionStatus := "Connected"
		if healthStatus == "red" {
			ingestionStatus = "Disconnected"
		} else if healthStatus == "yellow" {
			ingestionStatus = "Warning"
		}

		t.AppendRow([]interface{}{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for ingestion name (not implemented yet)
			ingestionStatus,
		})
	}

	if t.Length() == 0 { // Check the number of rows in the table
		logger.Info("No indices found")
		return
	}

	// Render the table
	t.Render()
}
