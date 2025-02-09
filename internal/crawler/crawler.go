package crawler

import (
	"net/url"
	"time"

	"crypto/md5"
	"encoding/hex"

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

var lastErrorTime time.Time

const errorLogCooldown = 10 * time.Second // Cooldown period for logging the same error

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger, log *logger.CustomLogger) (*Crawler, error) {
	// Create storage instance, which now initializes the Elasticsearch client
	storage, err := storage.NewStorage()
	if err != nil {
		return nil, err
	}

	// Initialize Colly collector with rate limiting and debugger
	collector := colly.NewCollector(
		colly.Async(true),
		// colly.Debugger(debugger),
	)

	// Parse the base URL to extract the domain
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err // Handle URL parsing error
	}
	allowedDomain := parsedURL.Hostname() // Extract the domain from the base URL

	// Set allowed domains to restrict crawling
	collector.AllowedDomains = []string{
		allowedDomain,
		"http://" + allowedDomain,
		"https://" + allowedDomain,
	}

	// Set rate limit
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       rateLimit,
	})

	// Set OnRequest callback
	collector.OnRequest(func(r *colly.Request) {
		startTime := time.Now() // Record the start time
		log.Info("Requesting URL", log.Field("url", r.URL.String()), log.Field("request_id", r.ID))

		// Log the time taken for the request
		defer func() {
			duration := time.Since(startTime)
			log.Info("Request completed", log.Field("url", r.URL.String()), log.Field("duration", duration))
		}()
	})

	// Set OnHTML callback to find and visit links
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if err := e.Request.Visit(link); err != nil {
			if err.Error() == "URL already visited" {
				log.Info("URL already visited", log.Field("link", link))
			} else if err.Error() == "Forbidden domain" {
				// Log as info instead of error
				log.Info("Forbidden domain", log.Field("link", link))
			} else if err.Error() == "Missing URL" {
				// Log as info instead of error
				log.Info("Missing URL", log.Field("link", link))
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

		// Log the content being indexed
		c.Logger.Info("Indexing document", c.Logger.Field("url", url), c.Logger.Field("content_length", len(content)), c.Logger.Field("content_preview", content[:100]))

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", c.Logger.Field("url", url))
			return
		}

		err := c.Storage.IndexDocument("example_index", docID, map[string]interface{}{"url": url, "content": content})
		if err != nil {
			if time.Since(lastErrorTime) > errorLogCooldown {
				c.Logger.Error("Error indexing document", c.Logger.Field("url", url), c.Logger.Field("error", err))
				lastErrorTime = time.Now()
			}
		} else {
			// Log success message for indexing
			c.Logger.Info("Successfully indexed document", c.Logger.Field("url", url), c.Logger.Field("docID", docID))
		}
	})

	err := c.Collector.Visit(url)
	if err != nil {
		c.Logger.Error("Error visiting URL", c.Logger.Field("url", url), c.Logger.Field("error", err))
	} else {
		// Log success message for visiting
		c.Logger.Info("Successfully visited URL", c.Logger.Field("url", url))
	}

	c.Collector.Wait()
}
