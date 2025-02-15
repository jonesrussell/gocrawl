package app

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the application module and its dependencies
var Module = fx.Module("app",
	fx.Provide(
		logger.NewLogger,
		// Provide a function that returns runCrawler
		func() func(ctx context.Context, storage storage.Interface) error {
			return runCrawler // Ensure runCrawler is provided correctly
		},
	),
	fx.Invoke(
		func(ctx context.Context, storage storage.Interface) error {
			log := logger.FromContext(ctx)
			log.Debug("Invoking runCrawler with provided context and storage")
			return runCrawler(ctx, storage)
		},
	),
)
