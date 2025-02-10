package app

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

// Module provides the application module and its dependencies
var Module = fx.Module("app",
	fx.Provide(
		NewLogger,
		// Provide the StartCrawler function
		func() func(ctx context.Context, cfg *config.Config) error {
			return StartCrawler
		},
	),
	// Invoke any initialization functions if needed
	fx.Invoke(
		runCrawler,
	),
)
