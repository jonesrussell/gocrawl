package app

import (
	"context"
	"fmt"

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

	// Merge CLI config with env config
	mergeConfigs(cfg, envConfig)

	app := newFxApp(envConfig)
	return app.Start(ctx)
}

func mergeConfigs(cliConfig, envConfig *config.Config) {
	if cliConfig.BaseURL != "" {
		envConfig.BaseURL = cliConfig.BaseURL
	}
	if cliConfig.MaxDepth != 0 {
		envConfig.MaxDepth = cliConfig.MaxDepth
	}
	if cliConfig.IndexName != "" {
		envConfig.IndexName = cliConfig.IndexName
	}
	if cliConfig.RateLimit != 0 {
		envConfig.RateLimit = cliConfig.RateLimit
	}
}

func newFxApp(cfg *config.Config) *fx.App {
	return fx.New(
		// Provide all dependencies
		fx.Provide(
			func() *config.Config { return cfg },
			NewLogger,
			storage.NewStorage,
			// Fix return type handling
			func(cfg *config.Config, log logger.Interface, store storage.Storage) (*crawler.Crawler, error) {
				params := crawler.Params{
					BaseURL:   cfg.BaseURL,
					MaxDepth:  cfg.MaxDepth,
					RateLimit: cfg.RateLimit,
					Debugger:  &logger.CollyDebugger{Logger: log},
					Logger:    log,
					Config:    cfg,
					Storage:   store,
				}
				result, err := crawler.NewCrawler(params)
				if err != nil {
					return nil, err
				}
				return result.Crawler, nil
			},
		),
		// Invoke the startup logic
		fx.Invoke(runCrawler),
	)
}

func runCrawler(
	lc fx.Lifecycle,
	c *crawler.Crawler,
	shutdowner fx.Shutdowner,
	log logger.Interface,
	cfg *config.Config,
) {
	log.Info("Starting crawler",
		"baseURL", cfg.BaseURL,
		"maxDepth", cfg.MaxDepth,
		"rateLimit", cfg.RateLimit)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := c.Start(ctx, shutdowner); err != nil {
					log.Error("Crawler failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return shutdowner.Shutdown()
		},
	})
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
