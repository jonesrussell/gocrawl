package main

import (
	"flag"

	"github.com/jonesrussell/gocrawl/internal/crawler" // Updated with the actual module path
	"github.com/jonesrussell/gocrawl/internal/logger"  // Import the logger package
	"go.uber.org/fx"
)

func main() {
	// Initialize the logger
	log, err := logger.NewLogger()
	if err != nil {
		panic(err) // Handle logger initialization error
	}

	// Define command-line flags for the URL and maxDepth
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 3, "The maximum depth to crawl") // New flag for maxDepth
	flag.Parse()                                                         // Parse the command-line flags

	app := fx.New(
		fx.Provide(func() (*crawler.Crawler, error) {
			return crawler.NewCrawler(*urlPtr, *maxDepthPtr) // Pass maxDepth from the flag
		}),
		fx.Invoke(func(c *crawler.Crawler) {
			content, err := c.Fetch(*urlPtr) // Fetch content from the URL
			if err != nil {
				log.Error("Error fetching content", log.Field("url", *urlPtr), log.Field("error", err))
				return
			}
			log.Info("Fetched content", log.Field("url", *urlPtr), log.Field("content", content))
			c.Start(*urlPtr) // Start crawling from this URL
		}),
	)

	app.Run()
}
