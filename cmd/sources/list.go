package sources

import (
	"context"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type listParams struct {
	ctx     context.Context
	sources *sources.Sources
	logger  common.Logger
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long:  `Display a list of all sources configured in sources.yml.`,
		Run:   runList,
	}
}

func runList(cmd *cobra.Command, _ []string) {
	var logger common.Logger

	app := fx.New(
		common.Module,
		fx.Invoke(func(s *sources.Sources, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				sources: s,
				logger:  l,
			}
			if err := executeList(params); err != nil {
				l.Error("Error executing list", "error", err)
				os.Exit(1)
			}
		}),
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
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
}

func executeList(p *listParams) error {
	common.PrintInfo("\nConfigured Sources")
	common.PrintDivider(17)
	common.PrintTableHeader("%-20s %-30s %-15s %-15s %-10s",
		"Name", "URL", "Article Index", "Content Index", "Max Depth")
	common.PrintDivider(92)

	for _, source := range p.sources.Sources {
		common.PrintTableHeader("%-20s %-30s %-15s %-15s %-10d",
			source.Name,
			source.URL,
			source.ArticleIndex,
			source.Index,
			source.MaxDepth)
	}

	common.PrintInfo("")
	return nil
}
