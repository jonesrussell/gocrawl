package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
)

// Constants for configuration
const (
	TimeoutDuration  = 5 * time.Second
	HTTPStatusOK     = 200
	DefaultRateLimit = time.Second
)

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	c.Logger.Debug("Starting crawl at base URL", "url", baseURL)

	// Log the entire configuration being used by the crawler
	c.Logger.Debug("Crawler configuration", "baseURL", baseURL)

	// Perform initial setup (e.g., test connection, ensure index)
	if err := c.Storage.TestConnection(ctx); err != nil {
		c.Logger.Error("Storage connection failed", "error", err)
		return fmt.Errorf("storage connection failed: %w", err)
	}

	if err := c.IndexSvc.EnsureIndex(ctx, c.IndexName); err != nil {
		c.Logger.Error("Index setup failed", "error", err)
		return fmt.Errorf("index setup failed: %w", err)
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

// ProcessPage handles article extraction
func (c *Crawler) ProcessPage(e *colly.HTMLElement) {
	c.Logger.Debug("Processing page", "url", e.Request.URL.String())
	article := c.ArticleService.ExtractArticle(e)
	if article == nil {
		c.Logger.Debug("No article extracted", "url", e.Request.URL.String())
		return
	}
	c.Logger.Debug("Article extracted", "url", e.Request.URL.String(), "title", article.Title)

	// Use the dynamic index name from the Crawler instance
	if err := c.Storage.IndexDocument(context.Background(), c.IndexName, article.ID, article); err != nil {
		c.Logger.Error("Failed to index article", "articleID", article.ID, "error", err)
		return
	}

	c.articleChan <- article
}

// Add these methods to the Crawler struct
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.Collector = collector
}

func (c *Crawler) SetService(svc article.Interface) {
	c.ArticleService = svc
}

// Getter methods for configuration
func (c *Crawler) GetBaseURL() string {
	return c.Config.Crawler.BaseURL
}

func (c *Crawler) GetMaxDepth() int {
	return c.Config.Crawler.MaxDepth
}

func (c *Crawler) GetRateLimit() time.Duration {
	return c.Config.Crawler.RateLimit
}

// ConfigureCollectorCallbacks sets up the callbacks for the Colly collector
func configureCollectorCallbacks(c *colly.Collector, crawler *Crawler) {
	c.OnRequest(func(r *colly.Request) {
		crawler.Logger.Debug("Requesting URL", "url", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.Logger.Debug("Received response", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error scraping", "url", r.Request.URL.String(), "error", err)
	})

	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		crawler.Logger.Debug("Found details", "url", e.Request.URL.String())
		crawler.ProcessPage(e) // Call ProcessPage on the Crawler instance
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}
		crawler.Logger.Debug("Found link", "url", link)
		if err := e.Request.Visit(link); err != nil {
			crawler.Logger.Debug("Could not visit link", "url", link, "error", err)
		}
	})

	if crawler.Debugger != nil {
		c.SetDebugger(crawler.Debugger)
	}
}
