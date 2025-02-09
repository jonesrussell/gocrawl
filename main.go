package main

import (
	"context"
	"flag"
	"os" // Import os for environment variable access
	"time"

	"github.com/joho/godotenv" // Import godotenv for loading .env files
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"

	"go.uber.org/fx"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file") // Handle error if .env file is not found
	}

	appEnv := os.Getenv("APP_ENV")     // Get the APP_ENV variable
	appDebug := os.Getenv("APP_DEBUG") // Get the APP_DEBUG variable

	// Initialize the logger based on the environment
	var log *logger.CustomLogger
	var err error
	if appEnv == "development" {
		log, err = logger.NewDevelopmentLogger() // Use development logger
	} else {
		log, err = logger.NewCustomLogger() // Use production logger
	}
	if err != nil {
		panic(err) // Handle logger initialization error
	}

	// Define command-line flags for the URL, maxDepth, and rateLimit
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests") // Set a longer rate limit for testing
	flag.Parse()                                                                             // Parse the command-line flags

	// Print configuration if APP_DEBUG is true
	if appDebug == "true" {
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
