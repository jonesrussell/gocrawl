package sources

import (
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the sources module
var Module = fx.Module("sources",
	fx.Provide(
		// Provide the sources configuration with crawler
		func(logger logger.Interface, c crawler.Interface) (*Sources, error) {
			sources, err := Load("sources.yml")
			if err != nil {
				return nil, err
			}
			sources.Logger = logger
			sources.Crawler = c
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
