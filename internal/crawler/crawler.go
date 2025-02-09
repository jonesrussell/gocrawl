package crawler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
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
	Done      chan bool
	IndexName string
}

var lastErrorTime time.Time

const errorLogCooldown = 10 * time.Second // Cooldown period for logging the same error

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger, log *logger.CustomLogger, cfg *config.Config) (*Crawler, error) {
	// Initialize storage with the config
	storage, err := initializeStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	log.Info("Successfully connected to Elasticsearch")

	// Initialize collector
	collector, err := initializeCollector(baseURL, maxDepth, rateLimit, debugger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize collector: %w", err)
	}

	// Configure logging for the collector
	configureCollectorLogging(collector, log)

	return &Crawler{
		BaseURL:   baseURL,
		Storage:   storage,
		MaxDepth:  maxDepth,
		RateLimit: rateLimit,
		Collector: collector,
		Logger:    log,
		Done:      make(chan bool),
		IndexName: cfg.IndexName,
	}, nil
}

func initializeStorage(cfg *config.Config, log *logger.CustomLogger) (*storage.Storage, error) {
	storage, err := storage.NewStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Test the connection to Elasticsearch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = storage.TestConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error testing connection: %w", err)
	}

	return storage, nil
}

func initializeCollector(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger) (*colly.Collector, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL: %w", err)
	}
	allowedDomain := parsedURL.Hostname()

	collector := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(maxDepth),
		colly.Debugger(debugger),
		// Set allowed domains directly
		colly.AllowedDomains(
			allowedDomain,
			"http://"+allowedDomain,
			"https://"+allowedDomain,
		),
	)

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       rateLimit,
	})

	return collector, nil
}

func configureCollectorLogging(collector *colly.Collector, log *logger.CustomLogger) {
	collector.OnRequest(func(r *colly.Request) {
		startTime := time.Now()
		logRequest(log, "Requesting URL", r, startTime)

		defer func() {
			duration := time.Since(startTime)
			logRequest(log, "Request completed", r, duration)
		}()
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if err := e.Request.Visit(link); err != nil {
			logVisitError(log, link, err)
		}
	})
}

func logRequest(log *logger.CustomLogger, message string, r *colly.Request, data interface{}) {
	log.Info(message, log.Field("url", r.URL.String()), log.Field("request_id", r.ID), log.Field("data", data))
}

func logVisitError(log *logger.CustomLogger, link string, err error) {
	switch err.Error() {
	case "URL already visited":
		log.Info("URL already visited", log.Field("link", link))
	case "Forbidden domain", "Missing URL":
		log.Info(err.Error(), log.Field("link", link))
	case "Max depth limit reached":
		// Log a simple message without stack trace
		log.Warn("Max depth limit reached", log.Field("link", link))
	default:
		log.Error("Error visiting link", log.Field("link", link), log.Field("error", err))
	}
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, url string) {
	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Check if the context is done before proceeding
		if ctx.Err() != nil {
			c.Logger.Warn("Crawling stopped due to context cancellation")
			return
		}

		content := e.Text
		docID := generateDocumentID(e.Request.URL.String())

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", c.Logger.Field("url", e.Request.URL.String()))
			return
		}

		c.indexDocument(ctx, c.IndexName, e.Request.URL.String(), content, docID)
	})

	err := c.Collector.Visit(url)
	if err != nil {
		c.Logger.Error("Error visiting URL", c.Logger.Field("url", url), c.Logger.Field("error", err))
	} else {
		c.Logger.Info("Successfully visited URL", c.Logger.Field("url", url))
	}

	c.Collector.Wait()

	// Signal that crawling is done
	c.Done <- true
}

func generateDocumentID(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Crawler) indexDocument(ctx context.Context, indexName, url, content, docID string) {
	// Log a concise summary instead of detailed content
	c.Logger.Debug("Indexing document",
		c.Logger.Field("url", url),
		c.Logger.Field("docID", docID))

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
