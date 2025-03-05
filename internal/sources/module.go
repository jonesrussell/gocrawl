package sources

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the sources module
var Module = fx.Module("sources",
	fx.Provide(
		// Provide the sources configuration without crawler dependency
		func(logger logger.Interface) (*Sources, error) {
			sources, err := Load("sources.yml")
			if err != nil {
				return nil, err
			}
			sources.Logger = logger
			return sources, nil
		},
		// Provide individual source configs
		fx.Annotate(
			func(s *Sources) []Config {
				return s.Sources
			},
			fx.ResultTags(`group:"sources"`),
		),
	),
)
