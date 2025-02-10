package app

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

// appShutdowner implements fx.Shutdowner
type appShutdowner struct {
	app *fx.App
}

func (s *appShutdowner) Shutdown(opts ...fx.ShutdownOption) error {
	return s.app.Stop(context.Background())
}

// StartCrawler initializes and starts the crawler
func StartCrawler(ctx context.Context, cfg *config.Config) error {
	log, err := logger.NewCustomLogger(logger.Params{
		Debug: true,
		Level: zapcore.DebugLevel,
	})
	if err != nil {
		return err
	}

	store, err := storage.NewStorage(cfg, log)
	if err != nil {
		return err
	}

	crawlerParams := crawler.Params{
		BaseURL:   cfg.CrawlerConfig.BaseURL,
		MaxDepth:  cfg.CrawlerConfig.MaxDepth,
		RateLimit: cfg.CrawlerConfig.RateLimit,
		Debugger:  &logger.CollyDebugger{},
		Logger:    log,
		Config:    cfg,
		Storage:   store.Storage,
	}

	crawlerResult, err := crawler.NewCrawler(crawlerParams)
	if err != nil {
		return err
	}

	var shutdowner *appShutdowner

	app := fx.New(
		fx.Supply(crawlerResult.Crawler),
		fx.Invoke(func(c *crawler.Crawler) {
			go func() {
				<-ctx.Done()
				time.Sleep(time.Second) // Give time for cleanup
			}()
		}),
	)

	shutdowner = &appShutdowner{app: app}

	if err := app.Start(ctx); err != nil {
		return err
	}

	return crawlerResult.Crawler.Start(ctx, shutdowner)
}

// Search performs a search query
func Search(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	log, err := logger.NewCustomLogger(logger.Params{
		Debug: true,
		Level: zapcore.InfoLevel,
	})
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{
		IndexName: index,
		CrawlerConfig: config.CrawlerConfig{
			BaseURL:   "",
			MaxDepth:  0,
			RateLimit: 0,
		},
	}

	store, err := storage.NewStorage(cfg, log)
	if err != nil {
		return nil, err
	}

	return store.Storage.Search(ctx, index, query)
}
