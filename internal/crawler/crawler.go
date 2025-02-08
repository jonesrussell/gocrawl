package crawler

import (
	"time"

	"crypto/md5"
	"encoding/hex"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly/v2"
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
}

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger, log *logger.CustomLogger) (*Crawler, error) {
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	storage := storage.NewStorage(esClient)

	// Initialize Colly collector with rate limiting and debugger
	collector := colly.NewCollector(
		colly.Async(true),
		colly.Debugger(debugger),
	)

	// Set allowed domains to restrict crawling
	collector.AllowedDomains = []string{baseURL} // Restrict to the base URL domain

	// Set rate limit
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       rateLimit,
	})

	// Set OnRequest callback
	collector.OnRequest(func(r *colly.Request) {
		log.Info("Requesting URL", log.Field("url", r.URL.String()), log.Field("request_id", r.ID))
	})

	// Set OnHTML callback to find and visit links
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if err := e.Request.Visit(link); err != nil {
			if err.Error() == "URL already visited" {
				log.Info("URL already visited", log.Field("link", link))
			} else {
				log.Error("Error visiting link", log.Field("link", link), log.Field("error", err))
			}
		}
	})

	return &Crawler{BaseURL: baseURL, Storage: storage, MaxDepth: maxDepth, RateLimit: rateLimit, Collector: collector, Logger: log}, nil
}

// Start method to begin crawling and indexing
func (c *Crawler) Start(url string) {
	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		content := e.Text

		hash := md5.Sum([]byte(url))
		docID := hex.EncodeToString(hash[:])

		err := c.Storage.IndexDocument("example_index", docID, map[string]interface{}{"url": url, "content": content})
		if err != nil {
			c.Logger.Error("Error indexing document", c.Logger.Field("url", url), c.Logger.Field("error", err))
		}
	})

	err := c.Collector.Visit(url)
	if err != nil {
		c.Logger.Error("Error visiting URL", c.Logger.Field("url", url), c.Logger.Field("error", err))
	}

	c.Collector.Wait()
}
