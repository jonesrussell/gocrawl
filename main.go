package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config" // Import the config package
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"

	"go.uber.org/fx"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	// Initialize the logger based on the environment
	var log *logger.CustomLogger
	if cfg.AppEnv == "development" {
		log, err = logger.NewDevelopmentLogger() // Use development logger
	} else {
		log, err = logger.NewCustomLogger() // Use production logger
	}
	if err != nil {
		log.Fatalf("Error initializing logger: %s", err)
	}

	// Define command-line flags for the URL, maxDepth, and rateLimit
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests")
	flag.Parse() // Parse the command-line flags

	// Print configuration if APP_DEBUG is true
	if cfg.AppDebug {
		log.Info("Crawling Configuration",
			log.Field("url", *urlPtr),
			log.Field("maxDepth", *maxDepthPtr),
			log.Field("rateLimit", rateLimitPtr.String()))
	}

	app := fx.New(
		fx.Provide(func(lc fx.Lifecycle) (*crawler.Crawler, error) {
			// Create a new Debugger
			debugger := logger.NewCustomDebugger(log)

			// Create a new Crawler
			crawlerInstance, err := crawler.NewCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, debugger, log)
			if err != nil {
				return nil, err
			}

			// Register a shutdown function to stop the crawler
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					crawlerInstance.Collector.Wait() // Wait for all requests to finish
					return nil
				},
			})

			return crawlerInstance, nil
		}),
		fx.Invoke(func(c *crawler.Crawler) {
			c.Start(*urlPtr) // Directly call Start to handle crawling and indexing
		}),
	)

	app.Run()
}
