package crawler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
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

var lastErrorTime time.Time

const errorLogCooldown = 10 * time.Second

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger, log *logger.CustomLogger, cfg *config.Config) (*Crawler, error) {
	storage, err := initializeStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	log.Info("Successfully connected to Elasticsearch")

	collectorInstance, err := collector.New(baseURL, maxDepth, rateLimit, debugger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize collector: %w", err)
	}

	collector.ConfigureLogging(collectorInstance, log)

	return &Crawler{
		BaseURL:   baseURL,
		Storage:   storage,
		MaxDepth:  maxDepth,
		RateLimit: rateLimit,
		Collector: collectorInstance,
		Logger:    log,
		IndexName: cfg.IndexName,
	}, nil
}

func initializeStorage(cfg *config.Config, log *logger.CustomLogger) (*storage.Storage, error) {
	storage, err := storage.NewStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = storage.TestConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error testing connection: %w", err)
	}

	return storage, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context) error {
	c.Logger.Info("Starting crawling process")

	c.Collector.OnRequest(func(r *colly.Request) {
		c.Logger.Debug("Requesting URL", c.Logger.Field("url", r.URL.String()))
	})

	c.Collector.OnResponse(func(r *colly.Response) {
		c.Logger.Debug("Received response", c.Logger.Field("url", r.Request.URL.String()), c.Logger.Field("status", r.StatusCode))
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

	if err := c.Collector.Visit(c.BaseURL); err != nil {
		return fmt.Errorf("error visiting URL: %w", err)
	}

	<-ctx.Done()
	return nil
}

func generateDocumentID(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Crawler) indexDocument(ctx context.Context, indexName, url, content, docID string) {
	c.Logger.Debug("Preparing to index document", c.Logger.Field("url", url), c.Logger.Field("docID", docID)) // Log before indexing

	err := c.Storage.IndexDocument(ctx, indexName, docID, map[string]interface{}{"url": url, "content": content})
	if err != nil {
		if time.Since(lastErrorTime) > errorLogCooldown {
			c.Logger.Error("Error indexing document", c.Logger.Field("url", url), c.Logger.Field("error", err))
			lastErrorTime = time.Now()
		}
	} else {
		c.Logger.Info("Successfully indexed document", c.Logger.Field("url", url), c.Logger.Field("docID", docID))
	}
}
