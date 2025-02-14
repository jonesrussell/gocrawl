package crawler

import (
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
	return cfg.Crawler.MaxDepth
}

func provideRateLimit(cfg *config.Config) time.Duration {
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
)
