package crawler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Constants for configuration
const (
	TimeoutDuration  = 5 * time.Second
	HTTPStatusOK     = 200
	DefaultRateLimit = time.Second
)

// Crawler represents a web crawler
type Crawler struct {
	Storage        storage.Interface
	Collector      *colly.Collector
	Logger         logger.Interface
	Debugger       *logger.CollyDebugger
	IndexName      string
	articleChan    chan *models.Article
	ArticleService article.Interface
	IndexSvc       storage.Interface
	Config         *config.Config
}

// Ensure Crawler implements the Interface
var _ Interface = (*Crawler)(nil)

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		return errors.New("base URL cannot be empty")
	}
	c.Logger.Debug("Starting crawl at base URL", "url", baseURL)

	// Log the entire configuration being used by the crawler
	c.Logger.Debug("Crawler configuration", "baseURL", baseURL)

	// Test storage connection before starting
	if err := c.Storage.Ping(ctx); err != nil {
		c.Logger.Error("Storage connection failed", "error", err)
		return fmt.Errorf("storage connection failed: %w", err)
	}

	// Create index with default mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "keyword",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	// Try to create the index - ignore error if it already exists
	if err := c.Storage.CreateIndex(ctx, c.IndexName, mapping); err != nil {
		c.Logger.Debug("Index creation failed (might already exist)", "error", err)
	}

	// Create a channel to track completion
	done := make(chan struct{})

	// Start crawling in a goroutine
	go func() {
		defer close(done)
		// Visit the base URL to start crawling
		if err := c.Collector.Visit(baseURL); err != nil {
			c.Logger.Error("Failed to visit base URL", "error", err)
			return
		}
		// Wait for collector to finish all requests
		c.Collector.Wait()
		c.Logger.Info("Crawler finished - no more links to visit")
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		c.Logger.Info("Crawler stopping due to context cancellation")
		return ctx.Err()
	case <-done:
		c.Logger.Info("Crawler completed successfully")
	}

	return nil
}

// Stop method to cleanly shut down the crawler
func (c *Crawler) Stop() {
	// Perform any necessary cleanup here
}

// SetCollector sets the collector for the crawler
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.Collector = collector
}

// SetService sets the article service for the crawler
func (c *Crawler) SetService(svc article.Interface) {
	c.ArticleService = svc
}

// GetBaseURL returns the base URL from the configuration
func (c *Crawler) GetBaseURL() string {
	return c.Config.Crawler.BaseURL
}
