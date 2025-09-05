// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("cmd_sources",
	// Include required modules
	common.Module,

	// Provide the command registrar
	fx.Provide(
		NewTableRenderer,
		NewLister,
		fx.Annotated{
			Group: "commands",
			Target: func(
				deps common.CommandDeps,
				lister *Lister,
				renderer *TableRenderer,
			) common.CommandRegistrar {
				return func(parent *cobra.Command) {
					cmd := &cobra.Command{
						Use:   "sources",
						Short: "Manage content sources",
						Long:  `Manage content sources for crawling`,
					}

					// Add subcommands
					cmd.AddCommand(
						&cobra.Command{
							Use:   "list",
							Short: "List all sources",
							RunE: func(cmd *cobra.Command, args []string) error {
								return lister.Start(cmd.Context())
							},
						},
					)

					parent.AddCommand(cmd)
				}
			},
		},
	),
)
