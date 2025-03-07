// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading, validation, and crawling operations through a
// YAML-based configuration system.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
// It configures the sources package as a dependency injection module using fx,
// providing both the main Sources instance and individual source configurations.
//
// The module provides:
// 1. A Sources instance loaded from YAML configuration
// 2. Individual source configurations tagged for group injection
//
// The module is designed to be used with the fx dependency injection framework
// and integrates with the main application's dependency graph.
var Module = fx.Module("sources",
	fx.Provide(
		// Provide the sources configuration without crawler dependency.
		// This constructor loads the sources from YAML and sets up the logger.
		// It is designed to be called early in the application lifecycle,
		// before the crawler is available.
		func(logger logger.Interface) (*Sources, error) {
			sources, err := Load("sources.yml")
			if err != nil {
				return nil, err
			}
			sources.Logger = logger
			return sources, nil
		},
		// Provide individual source configs for group injection.
		// This allows other parts of the application to receive individual
		// source configurations tagged with the "sources" group.
		// This is useful for components that need to work with specific sources
		// or need to process sources independently.
		fx.Annotate(
			func(s *Sources) []Config {
				return s.Sources
			},
			fx.ResultTags(`group:"sources"`),
		),
	),
)
