package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"

	"go.uber.org/fx"
)

func main() {
	// Define command-line flags
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests")
	flag.Parse() // Parse the command-line flags

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	// Initialize the logger based on the environment
	loggerInstance := newLogger(cfg)

	// Print configuration if APP_DEBUG is true
	if cfg.AppDebug {
		loggerInstance.Info("Crawling Configuration",
			loggerInstance.Field("url", *urlPtr),
			loggerInstance.Field("maxDepth", *maxDepthPtr),
			loggerInstance.Field("rateLimit", rateLimitPtr.String()))
	}

	// Create and run the Fx application
	app := fx.New(
		fx.Provide(
			// Provide the logger
			func() *logger.CustomLogger {
				return loggerInstance
			},
			func(lc fx.Lifecycle) (*crawler.Crawler, error) {
				return initializeCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, loggerInstance, cfg, lc)
			},
		),
		fx.Invoke(func(c *crawler.Crawler) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			c.Start(ctx, *urlPtr) // Directly call Start to handle crawling and indexing
		}),
	)

	app.Run()
}

func newLogger(cfg *config.Config) *logger.CustomLogger {
	if cfg.AppEnv == "development" {
		loggerInstance, err := logger.NewDevelopmentLogger() // Use development logger
		if err != nil {
			log.Fatalf("Error initializing development logger: %s", err)
		}
		return loggerInstance
	}

	loggerInstance, err := logger.NewCustomLogger() // Use production logger
	if err != nil {
		log.Fatalf("Error initializing production logger: %s", err)
	}
	return loggerInstance
}

func initializeCrawler(url string, maxDepth int, rateLimit time.Duration, loggerInstance *logger.CustomLogger, cfg *config.Config, lc fx.Lifecycle) (*crawler.Crawler, error) {
	// Create a new Debugger
	debugger := logger.NewCustomDebugger(loggerInstance)

	// Create a new Crawler with the config
	crawlerInstance, err := crawler.NewCrawler(url, maxDepth, rateLimit, debugger, loggerInstance, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crawler: %w", err)
	}

	// Register a shutdown function to stop the crawler
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			crawlerInstance.Collector.Wait() // Wait for all requests to finish
			return nil
		},
	})

	return crawlerInstance, nil
}
