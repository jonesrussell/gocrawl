// Package crawler provides the core crawling functionality for the application.
// It manages the crawling process, coordinates between components, and handles
// configuration and error management.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Constants for configuration
const (
	// TimeoutDuration is the default timeout for operations
	TimeoutDuration = 5 * time.Second
	// HTTPStatusOK represents a successful HTTP response
	HTTPStatusOK = 200
	// DefaultRateLimit is the default rate limit for requests
	DefaultRateLimit = time.Second
)

// Crawler represents a web crawler instance that manages the crawling process.
// It coordinates between the collector, storage, and logger components while
// handling configuration and error management.
type Crawler struct {
	// Storage handles content storage operations
	Storage storage.Interface
	// Collector manages the actual web page collection
	Collector *colly.Collector
	// Logger provides structured logging capabilities
	Logger logger.Interface
	// Debugger handles debugging operations
	Debugger *logger.CollyDebugger
	// IndexName is the name of the Elasticsearch index
	IndexName string
	// articleChan is a channel for processing articles
	articleChan chan *models.Article
	// ArticleService handles article-specific operations
	ArticleService article.Interface
	// IndexManager manages index operations
	IndexManager api.IndexManager
	// Config holds the crawler configuration
	Config *config.Config
	// ContentProcessor handles content processing
	ContentProcessor models.ContentProcessor
}

// Ensure Crawler implements the Interface
var _ Interface = (*Crawler)(nil)

// Start begins the crawling process at the specified base URL.
// It manages the crawling lifecycle, including setup, execution, and cleanup.
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

	// Ensure article index exists with default mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
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

	if err := c.IndexManager.EnsureIndex(ctx, c.IndexName, mapping); err != nil {
		c.Logger.Error("Failed to ensure article index exists", "error", err)
		return fmt.Errorf("failed to ensure article index exists: %w", err)
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
		c.Logger.Debug("Collector finished all requests")
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

// Stop performs cleanup operations when the crawler is stopped.
// It ensures all resources are properly released.
func (c *Crawler) Stop() {
	// Perform any necessary cleanup here
}

// SetCollector sets the collector for the crawler.
// This allows for dependency injection and testing.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.Collector = collector
}

// SetService sets the article service for the crawler.
// This allows for dependency injection and testing.
func (c *Crawler) SetService(svc article.Interface) {
	c.ArticleService = svc
}

// GetBaseURL returns the base URL from the configuration.
// This is used for validation and logging purposes.
func (c *Crawler) GetBaseURL() string {
	return c.Config.Crawler.BaseURL
}

// GetIndexManager returns the index manager interface
func (c *Crawler) GetIndexManager() api.IndexManager {
	return c.IndexManager
}
