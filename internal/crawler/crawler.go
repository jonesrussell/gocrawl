package crawler

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"

	"github.com/gocolly/colly/v2"
)

// Crawler struct to hold configuration or state if needed
type Crawler struct {
	BaseURL   string
	Storage   *storage.Storage
	MaxDepth  int
	RateLimit time.Duration
	Collector *colly.Collector
	Logger    *logger.CustomLogger
}

var lastErrorTime time.Time

const errorLogCooldown = 10 * time.Second // Cooldown period for logging the same error

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger, log *logger.CustomLogger, cfg *config.Config) (*Crawler, error) {
	// Initialize storage with the config
	storage, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}

	// Test the connection to Elasticsearch
	err = storage.TestConnection()
	if err != nil {
		log.Fatalf("Error testing connection: %s", err) // Exit the application on connection failure
	}

	log.Info("Successfully connected to Elasticsearch")

	// Initialize collector
	collector, err := initializeCollector(baseURL, maxDepth, rateLimit, debugger)
	if err != nil {
		return nil, err
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
	}, nil
}

func initializeCollector(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger) (*colly.Collector, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
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
		link := e.Attr("href")
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
	default:
		log.Error("Error visiting link", log.Field("link", link), log.Field("error", err))
	}
}

func (c *Crawler) Start(url string) {
	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		content := e.Text
		docID := generateDocumentID(url)

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", c.Logger.Field("url", url))
			return
		}

		c.indexDocument(url, content, docID)
	})

	err := c.Collector.Visit(url)
	if err != nil {
		c.Logger.Error("Error visiting URL", c.Logger.Field("url", url), c.Logger.Field("error", err))
	} else {
		c.Logger.Info("Successfully visited URL", c.Logger.Field("url", url))
	}

	c.Collector.Wait()
}

func generateDocumentID(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Crawler) indexDocument(url, content, docID string) {
	c.Logger.Info("Indexing document", c.Logger.Field("url", url), c.Logger.Field("content_length", len(content)), c.Logger.Field("content_preview", content[:100]))

	err := c.Storage.IndexDocument("example_index", docID, map[string]interface{}{"url": url, "content": content})
	if err != nil {
		if time.Since(lastErrorTime) > errorLogCooldown {
			c.Logger.Error("Error indexing document", c.Logger.Field("url", url), c.Logger.Field("error", err))
			lastErrorTime = time.Now()
		}
	} else {
		c.Logger.Info("Successfully indexed document", c.Logger.Field("url", url), c.Logger.Field("docID", docID))
	}
}
