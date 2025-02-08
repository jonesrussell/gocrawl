package main

import (
	"context"
	"flag"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler" // Updated with the actual module path
	"github.com/jonesrussell/gocrawl/internal/logger"  // Import the logger package

	// Import Colly debug package
	"go.uber.org/fx"
)

func main() {
	// Initialize the logger
	log, err := logger.NewLogger()
	if err != nil {
		panic(err) // Handle logger initialization error
	}

	// Define command-line flags for the URL, maxDepth, and rateLimit
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")                     // New flag for maxDepth
	rateLimitPtr := flag.Duration("rateLimit", 2*time.Second, "Rate limit between requests") // New flag for rate limit
	flag.Parse()                                                                             // Parse the command-line flags

	app := fx.New(
		fx.Provide(func(lc fx.Lifecycle) (*crawler.Crawler, error) {
			// Create a new Debugger
			debugger := logger.NewCustomDebugger(log)

			// Create a new Crawler
			crawlerInstance, err := crawler.NewCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, debugger)
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
