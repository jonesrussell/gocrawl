package crawler

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Provide named values for the crawler
func provideBaseURL(cfg *config.Config) string {
	return cfg.Crawler.BaseURL
}

func provideMaxDepth(cfg *config.Config) int {
	if cfg.Crawler.MaxDepth <= 0 {
		return DefaultMaxDepth
	}
	return cfg.Crawler.MaxDepth
}

func provideRateLimit(cfg *config.Config) time.Duration {
	if cfg.Crawler.RateLimit <= 0 {
		return DefaultRateLimit
	}
	return cfg.Crawler.RateLimit
}

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// Module provides the crawler module and its dependencies
var Module = fx.Module("crawler",
	fx.Provide(
		fx.Annotated{
			Name:   "baseURL",
			Target: provideBaseURL,
		},
		fx.Annotated{
			Name:   "maxDepth",
			Target: provideMaxDepth,
		},
		fx.Annotated{
			Name:   "rateLimit",
			Target: provideRateLimit,
		},
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
