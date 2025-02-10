package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Constants for configuration
const (
	TimeoutDuration = 5 * time.Second
	HTTPStatusOK    = 200
)

// Crawler struct to hold configuration or state if needed
type Crawler struct {
	BaseURL   string
	Storage   storage.Storage
	MaxDepth  int
	RateLimit time.Duration
	Collector *colly.Collector
	Logger    logger.Interface
	IndexName string
}

// Params holds the parameters for creating a Crawler
type Params struct {
	fx.In

	BaseURL   string        `name:"baseURL"`
	MaxDepth  int           `name:"maxDepth"`
	RateLimit time.Duration `name:"rateLimit"`
	Debugger  *logger.CollyDebugger
	Logger    logger.Interface
	Config    *config.Config
	Storage   storage.Storage
}

// Result holds the dependencies for the crawler
type Result struct {
	fx.Out

	Crawler *Crawler
}

// NewCrawler initializes a new Crawler
func NewCrawler(p Params) (Result, error) {
	// Use the provided storage instance
	storageInstance := p.Storage

	p.Logger.Info("Successfully connected to Elasticsearch")

	collectorResult, err := collector.New(collector.Params{
		BaseURL:   p.BaseURL,
		MaxDepth:  p.MaxDepth,
		RateLimit: p.RateLimit,
		Debugger:  p.Debugger,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to initialize collector: %w", err)
	}
	collectorInstance := collectorResult.Collector // Extract the Collector

	collector.ConfigureLogging(collectorInstance, p.Logger)

	return Result{Crawler: &Crawler{
		BaseURL:   p.BaseURL,
		Storage:   storageInstance,
		MaxDepth:  p.MaxDepth,
		RateLimit: p.RateLimit,
		Collector: collectorInstance,
		Logger:    p.Logger,
		IndexName: p.Config.IndexName,
	}}, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, shutdowner fx.Shutdowner) error {
	// Test the connection before starting the crawl
	if err := c.Storage.TestConnection(ctx); err != nil {
		c.Logger.Error("Failed to connect to storage", err)
		return err
	}

	c.Logger.Info("Starting crawling process")
	c.configureCollectors(ctx)

	if err := c.Collector.Visit(c.BaseURL); err != nil {
		c.Logger.Error("Error visiting URL", err)
		return err
	}

	// Wait for all requests to finish
	c.Logger.Info("Crawling process finished, waiting for all requests to complete...")
	c.Collector.Wait() // Wait for all requests to finish
	c.Logger.Info("All requests completed, initiating shutdown...")
	if err := shutdowner.Shutdown(); err != nil {
		c.Logger.Error("Error during shutdown", err)
	}

	return nil
}

// Helper method to configure collectors
func (c *Crawler) configureCollectors(ctx context.Context) {
	c.Collector.OnRequest(func(r *colly.Request) {
		c.Logger.Debug("Requesting URL", r.URL.String())
	})

	c.Collector.OnResponse(func(r *colly.Response) {
		c.Logger.Debug("Received response", r.Request.URL.String(), r.StatusCode)
		if r.StatusCode != HTTPStatusOK {
			c.Logger.Warn(
				"Non-200 response received",
				r.Request.URL.String(),
				r.StatusCode,
			)
		}
	})

	c.Collector.OnError(func(r *colly.Response, err error) {
		c.Logger.Error("Error occurred", r.Request.URL.String(), err)
	})

	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			c.Logger.Warn("Crawling stopped due to context cancellation", ctx.Err())
			return
		}

		content := e.Text
		docID := generateDocumentID(e.Request.URL.String())

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", e.Request.URL.String())
			return
		}

		c.Logger.Debug("Indexing document", e.Request.URL.String(), docID)
		c.indexDocument(ctx, c.IndexName, e.Request.URL.String(), content, docID)
	})
}

func generateDocumentID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Crawler) indexDocument(ctx context.Context, indexName, url, content, docID string) {
	c.Logger.Debug(
		"Preparing to index document",
		url,
		docID,
	) // Log before indexing

	err := c.Storage.IndexDocument(ctx, indexName, docID, map[string]interface{}{"url": url, "content": content})
	if err != nil {
		c.Logger.Error("Error indexing document", url, err)
	} else {
		c.Logger.Info("Successfully indexed document", url, docID)
	}
}
