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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// Load .env file if it exists
	godotenv.Load()
}

// StartCrawler initializes and starts the crawler
func StartCrawler(ctx context.Context, cfg *config.Config) error {
	// Load base config from environment
	envConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI parameters
	if cfg.BaseURL != "" {
		envConfig.BaseURL = cfg.BaseURL
	}
	if cfg.MaxDepth != 0 {
		envConfig.MaxDepth = cfg.MaxDepth
	}
	if cfg.IndexName != "" {
		envConfig.IndexName = cfg.IndexName
	}

	log, err := NewLogger(envConfig)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	store, err := storage.NewStorage(envConfig, log)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	crawlerParams := crawler.Params{
		BaseURL:   envConfig.BaseURL,
		MaxDepth:  envConfig.MaxDepth,
		RateLimit: time.Second,
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

	app := fx.New(
		fx.Supply(crawlerResult.Crawler),
		fx.Invoke(func(c *crawler.Crawler) {
			go func() {
				<-ctx.Done()
				time.Sleep(time.Second)
			}()
		}),
	)

	return app.Start(ctx)
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

func NewConfig() (*config.Config, error) {
	envConfig, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return envConfig, nil
}

func NewLogger(cfg *config.Config) (logger.Interface, error) {
	var zapConfig zap.Config
	if cfg.AppEnv == "development" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	return logger.NewCustomLogger(logger.Params{
		Debug:  cfg.AppEnv == "development",
		Level:  zapConfig.Level.Level(),
		AppEnv: cfg.AppEnv,
	})
}

func NewCrawlerParams(cfg *config.Config, log logger.Interface, store storage.Storage) crawler.Params {
	return crawler.Params{
		BaseURL:   cfg.BaseURL,
		MaxDepth:  cfg.MaxDepth,
		RateLimit: 1 * time.Second,                    // Default rate limit
		Debugger:  &logger.CollyDebugger{Logger: log}, // Use struct initialization
		Logger:    log,
		Config:    cfg,
		Storage:   store,
	}
}
