package indices

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// TableWidth is the total width of the table output
	TableWidth = 92
)

type listParams struct {
	ctx     context.Context
	storage common.Storage
	logger  common.Logger
}

// listCommand returns the list command
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		Run: runList,
	}
}

func runList(cmd *cobra.Command, _ []string) {
	var logger common.Logger
	var exitCode int

	app := fx.New(
		common.Module,
		fx.Invoke(func(s common.Storage, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				storage: s,
				logger:  l,
			}
			if err := executeList(params); err != nil {
				l.Error("Error executing list", "error", err)
				exitCode = 1
			}
		}),
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer func() {
		cancel()
		if err := app.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			if logger != nil {
				logger.Error("Error stopping application", "error", err)
				exitCode = 1
			}
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	if err := app.Start(ctx); err != nil {
		if logger != nil {
			logger.Error("Error starting application", "error", err)
		}
		exitCode = 1
		return
	}
}

func executeList(p *listParams) error {
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return err
	}

	// Filter out internal indices
	var filteredIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	if len(filteredIndices) == 0 {
		p.logger.Info("No indices found")
		return nil
	}

	return printIndices(p.ctx, filteredIndices, p.storage, p.logger)
}

func printIndices(ctx context.Context, indices []string, storage common.Storage, logger common.Logger) error {
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

		ingestionStatus := getIngestionStatus(healthStatus)

		t.AppendRow([]interface{}{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for ingestion name (not implemented yet)
			ingestionStatus,
		})
	}

	if t.Length() == 0 {
		logger.Info("No indices found")
		return nil
	}

	t.Render()
	return nil
}

func getIngestionStatus(healthStatus string) string {
	switch healthStatus {
	case "red":
		return "Disconnected"
	case "yellow":
		return "Warning"
	default:
		return "Connected"
	}
}
