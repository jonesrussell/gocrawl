// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"errors"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the search command dependencies
var Module = fx.Module("search",
	common.Module,
	fx.Provide(
		fx.Annotated{
			Group: "commands",
			Target: func(
				cfg config.Interface,
				log logger.Interface,
				storage types.Interface,
			) common.CommandRegistrar {
				return func(parent *cobra.Command) {
					cmd := &cobra.Command{
						Use:   "search [query]",
						Short: "Search content in Elasticsearch",
						Long:  `Search for content in Elasticsearch using the provided query`,
						Args:  cobra.ExactArgs(1),
						RunE: func(cmd *cobra.Command, args []string) error {
							// Validate Elasticsearch configuration
							esConfig := cfg.GetElasticsearchConfig()
							if esConfig == nil {
								return errors.New("elasticsearch configuration is required")
							}

							// TODO: Implement search functionality
							return nil
						},
					}

					parent.AddCommand(cmd)
				}
			},
		},
	),
)
