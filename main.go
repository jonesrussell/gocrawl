package main

import (
	"flag"

	"go.uber.org/fx"
	"github.com/jonesrussell/gocrawl/internal/crawler" // Updated with the actual module path
	"github.com/jonesrussell/gocrawl/internal/logger" // Import the logger package
)

func main() {
	// Initialize the logger
	log, err := logger.NewLogger()
	if err != nil {
		panic(err) // Handle logger initialization error
	}

	// Define a command-line flag for the URL
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	flag.Parse() // Parse the command-line flags

	app := fx.New(
		fx.Provide(func() (*crawler.Crawler, error) {
			return crawler.NewCrawler(*urlPtr) // Pass the URL from the flag
		}),
		fx.Invoke(func(c *crawler.Crawler) {
			content, err := c.Fetch(*urlPtr) // Fetch content from the URL
			if err != nil {
				log.Error("Error fetching content", logger.Field("url", *urlPtr), logger.Field("error", err))
				return
			}
			log.Info("Fetched content", logger.Field("url", *urlPtr), logger.Field("content", content))
			c.Start(*urlPtr) // Start crawling from this URL
		}),
	)

	app.Run()
}
