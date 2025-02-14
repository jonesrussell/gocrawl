package crawler

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// Module provides the crawler module and its dependencies
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		NewCrawler,
	),
	fx.Invoke(registerHooks),
)

// registerHooks sets up the lifecycle hooks for the crawler
func registerHooks(lc fx.Lifecycle, crawler *Crawler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return crawler.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			crawler.Stop()
			return nil
		},
	})
}
