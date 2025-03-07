package sources

import (
	"context"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// HeaderWidth is the width of the header divider
	HeaderWidth = 17
	// TableWidth is the width of the table divider
	TableWidth = 92
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
	var exitCode int

	app := fx.New(
		common.Module,
		fx.Invoke(func(s *sources.Sources, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				sources: s,
				logger:  l,
			}
			displaySourcesList(params)
		}),
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer func() {
		cancel()
		if err := app.Stop(ctx); err != nil {
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

func displaySourcesList(p *listParams) {
	common.PrintInfof("\nConfigured Sources")
	common.PrintDivider(HeaderWidth)
	common.PrintTableHeaderf("%-20s %-30s %-15s %-15s %-10s",
		"Name", "URL", "Article Index", "Content Index", "Max Depth")
	common.PrintDivider(TableWidth)

	for _, source := range p.sources.Sources {
		common.PrintTableHeaderf("%-20s %-30s %-15s %-15s %-10d",
			source.Name,
			source.URL,
			source.ArticleIndex,
			source.Index,
			source.MaxDepth)
	}

	common.PrintInfof("")
}
