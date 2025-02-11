package app

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// StartCrawler initializes and starts the crawler
func StartCrawler(ctx context.Context, cfg *config.Config) error {
	app := newFxApp(cfg)
	return app.Start(ctx)
}

func newFxApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			NewLogger,
			storage.NewStorage,
			func(cfg *config.Config, log logger.Interface, store storage.Storage) (*crawler.Crawler, error) {
				params := crawler.Params{
					BaseURL:   cfg.Crawler.BaseURL,
					MaxDepth:  cfg.Crawler.MaxDepth,
					RateLimit: cfg.Crawler.RateLimit,
					Debugger:  &logger.CollyDebugger{Logger: log},
					Logger:    log,
					Config:    cfg,
					Storage:   store,
				}
				result, err := crawler.NewCrawler(params)
				if err != nil {
					log.Error("Failed to create crawler", "error", err)
					return nil, err
				}
				log.Debug("Crawler created successfully", "baseURL", params.BaseURL)
				return result.Crawler, nil
			},
		),
		fx.Invoke(runCrawler),
	)
}

func runCrawler(
	lc fx.Lifecycle,
	c *crawler.Crawler,
	log logger.Interface,
	cfg *config.Config,
) {
	log.Info("Starting crawler",
		"baseURL", cfg.Crawler.BaseURL,
		"maxDepth", cfg.Crawler.MaxDepth,
		"rateLimit", cfg.Crawler.RateLimit)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Debug("Crawler is starting...")
			go func() {
				if err := c.Start(ctx); err != nil {
					log.Error("Crawler failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping crawler...")
			c.Stop()
			log.Debug("Crawler stopped successfully")
			return nil
		},
	})
}
