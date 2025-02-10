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
	Storage   *storage.Storage
	MaxDepth  int
	RateLimit time.Duration
	Collector *colly.Collector
	Logger    *logger.CustomLogger
	IndexName string
}

// Params holds the dependencies for creating a new Crawler
type Params struct {
	fx.In

	BaseURL   string        `name:"baseURL"`
	MaxDepth  int           `name:"maxDepth"`
	RateLimit time.Duration `name:"rateLimit"`
	Debugger  *logger.CollyDebugger
	Logger    *logger.CustomLogger
	Config    *config.Config
}

// Result holds the dependencies for the crawler
type Result struct {
	fx.Out

	Crawler *Crawler
}

// NewCrawler initializes a new Crawler
func NewCrawler(p Params) (Result, error) {
	storageResult, err := initializeStorage(p.Config, p.Logger)
	if err != nil {
		return Result{}, fmt.Errorf("failed to initialize storage: %w", err)
	}
	storageInstance := storageResult.Storage // Extract the Storage

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

func initializeStorage(cfg *config.Config, log *logger.CustomLogger) (*storage.Result, error) {
	storageResult, err := storage.NewStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDuration)
	defer cancel()

	err = storageResult.Storage.TestConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error testing connection: %w", err)
	}

	return &storageResult, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, shutdowner fx.Shutdowner) error {
	c.Logger.Info("Starting crawling process")
	c.configureCollectors(ctx)

	if err := c.Collector.Visit(c.BaseURL); err != nil {
		c.Logger.Error("Error visiting URL", c.Logger.Field("error", err))
		return err
	}

	// Wait for all requests to finish
	c.Logger.Info("Crawling process finished, waiting for all requests to complete...")
	c.Collector.Wait() // Wait for all requests to finish
	c.Logger.Info("All requests completed, initiating shutdown...")
	if err := shutdowner.Shutdown(); err != nil {
		c.Logger.Error("Error during shutdown", c.Logger.Field("error", err))
	}

	return nil
}

// Helper method to configure collectors
func (c *Crawler) configureCollectors(ctx context.Context) {
	c.Collector.OnRequest(func(r *colly.Request) {
		c.Logger.Debug("Requesting URL", c.Logger.Field("url", r.URL.String()))
	})

	c.Collector.OnResponse(func(r *colly.Response) {
		c.Logger.Debug(
			"Received response",
			c.Logger.Field("url", r.Request.URL.String()),
			c.Logger.Field("status", r.StatusCode),
		)
		if r.StatusCode != HTTPStatusOK {
			c.Logger.Warn(
				"Non-200 response received",
				c.Logger.Field("url", r.Request.URL.String()),
				c.Logger.Field("status", r.StatusCode),
			)
		}
	})

	c.Collector.OnError(func(r *colly.Response, err error) {
		c.Logger.Error("Error occurred", c.Logger.Field("url", r.Request.URL.String()), c.Logger.Field("error", err))
	})

	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			c.Logger.Warn("Crawling stopped due to context cancellation", c.Logger.Field("error", ctx.Err()))
			return
		}

		content := e.Text
		docID := generateDocumentID(e.Request.URL.String())

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", c.Logger.Field("url", e.Request.URL.String()))
			return
		}

		c.Logger.Debug("Indexing document", c.Logger.Field("url", e.Request.URL.String()), c.Logger.Field("docID", docID))
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
		c.Logger.Field("url", url),
		c.Logger.Field("docID", docID),
	) // Log before indexing

	err := c.Storage.IndexDocument(ctx, indexName, docID, map[string]interface{}{"url": url, "content": content})
	if err != nil {
		c.Logger.Error("Error indexing document", c.Logger.Field("url", url), c.Logger.Field("error", err))
	} else {
		c.Logger.Info("Successfully indexed document", c.Logger.Field("url", url), c.Logger.Field("docID", docID))
	}
}
