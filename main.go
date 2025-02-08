package main

import (
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
		fx.Provide(func() (*crawler.Crawler, error) {
			// Create a new Debugger
			debugger := logger.NewCustomDebugger(log)
			return crawler.NewCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, debugger) // Pass debugger to the crawler
		}),
		fx.Invoke(func(c *crawler.Crawler) {
			c.Start(*urlPtr) // Directly call Start to handle crawling and indexing
		}),
	)

	app.Run()
}
