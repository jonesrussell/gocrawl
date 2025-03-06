package sources

import (
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/shared"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// listCommand creates and returns the list subcommand
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long: `Display a list of all sources configured in sources.yml.
Shows details like URL, rate limit, and index names for each source.

Example:
  gocrawl sources list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := fx.New(
				fx.WithLogger(func() fxevent.Logger {
					return &fxevent.NopLogger
				}),
				fx.Provide(
					func() *config.Config {
						return shared.Config
					},
					func() logger.Interface {
						return shared.Logger
					},
				),
				sources.Module,
				fx.Invoke(func(s *sources.Sources) {
					fmt.Println("\nConfigured Sources")
					fmt.Println("-----------------")
					fmt.Printf("%-20s %-30s %-15s %-15s %-10s\n",
						"Name",
						"URL",
						"Article Index",
						"Content Index",
						"Max Depth")
					fmt.Println(strings.Repeat("-", 92))

					for _, source := range s.Sources {
						fmt.Printf("%-20s %-30s %-15s %-15s %-10d\n",
							source.Name,
							source.URL,
							source.ArticleIndex,
							source.Index,
							source.MaxDepth)
					}
					fmt.Println()
				}),
			)

			if err := app.Start(cmd.Context()); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			if err := app.Stop(cmd.Context()); err != nil {
				return fmt.Errorf("error stopping application: %w", err)
			}

			return nil
		},
	}
}
