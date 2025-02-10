package app

import (
	"context"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

func init() {
	// Load .env file if it exists
	godotenv.Load()
}

// appShutdowner implements fx.Shutdowner
type appShutdowner struct {
	app *fx.App
}

func (s *appShutdowner) Shutdown(opts ...fx.ShutdownOption) error {
	return s.app.Stop(context.Background())
}

// StartCrawler initializes and starts the crawler
func StartCrawler(ctx context.Context, cfg *config.Config) error {
	// Load full config from environment
	envConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge CLI config with environment config
	envConfig.CrawlerConfig = cfg.CrawlerConfig

	log, err := logger.NewCustomLogger(logger.Params{
		Debug:  envConfig.AppDebug,
		Level:  zapcore.DebugLevel,
		AppEnv: envConfig.AppEnv,
	})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	store, err := storage.NewStorage(envConfig, log)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	crawlerParams := crawler.Params{
		BaseURL:   envConfig.CrawlerConfig.BaseURL,
		MaxDepth:  envConfig.CrawlerConfig.MaxDepth,
		RateLimit: envConfig.CrawlerConfig.RateLimit,
		Debugger:  &logger.CollyDebugger{Logger: log},
		Logger:    log,
		Config:    envConfig,
		Storage:   store.Storage,
	}

	log.Info("Starting crawler",
		"baseURL", crawlerParams.BaseURL,
		"maxDepth", crawlerParams.MaxDepth,
		"rateLimit", crawlerParams.RateLimit)

	crawlerResult, err := crawler.NewCrawler(crawlerParams)
	if err != nil {
		return fmt.Errorf("failed to create crawler: %w", err)
	}

	var shutdowner *appShutdowner

	app := fx.New(
		fx.Supply(crawlerResult.Crawler),
		fx.Invoke(func(c *crawler.Crawler) {
			go func() {
				<-ctx.Done()
				time.Sleep(time.Second)
			}()
		}),
	)

	shutdowner = &appShutdowner{app: app}

	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return crawlerResult.Crawler.Start(ctx, shutdowner)
}

// Search performs a search query
func Search(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	// Load config from environment
	envConfig, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Override index name from CLI
	envConfig.IndexName = index

	log, err := logger.NewCustomLogger(logger.Params{
		Debug: true,
		Level: zapcore.InfoLevel,
	})
	if err != nil {
		return nil, err
	}

	store, err := storage.NewStorage(envConfig, log)
	if err != nil {
		return nil, err
	}

	return store.Storage.Search(ctx, index, query)
}
