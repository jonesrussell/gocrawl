package main

import (
	"context"
	"flag"
	"time"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests")
	flag.Parse()

	app := createApp(*urlPtr, *maxDepthPtr, *rateLimitPtr)

	// Run the application
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}

func createApp(url string, maxDepth int, rateLimit time.Duration) *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,
		fx.Provide(
			fx.Annotated{
				Name:   "baseURL",
				Target: func() string { return url },
			},
			fx.Annotated{
				Name:   "maxDepth",
				Target: func() int { return maxDepth },
			},
			fx.Annotated{
				Name:   "rateLimit",
				Target: func() time.Duration { return rateLimit },
			},
			func(l *logger.CustomLogger) *logger.CustomDebugger {
				return logger.NewCustomDebugger(l)
			},
		),
		fx.WithLogger(func(l *logger.CustomLogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l.GetZapLogger()}
		}),
		fx.Invoke(registerHooks),
	)
}

func registerHooks(lc fx.Lifecycle, c *crawler.Crawler, shutdowner fx.Shutdowner, log *logger.CustomLogger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting crawling process...")
			startCrawling(ctx, c, shutdowner, log)
			return nil
		},
	})
}

func startCrawling(ctx context.Context, c *crawler.Crawler, shutdowner fx.Shutdowner, log *logger.CustomLogger) {
	if err := c.Start(ctx, shutdowner); err != nil {
		log.Errorf("Error during crawling: %s", err)
	}
	log.Info("Crawling process finished")
}
