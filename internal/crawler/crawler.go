package crawler

import (
	"log"
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
}

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int, rateLimit time.Duration, debugger *logger.CustomDebugger) (*Crawler, error) {
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

	// Set rate limit
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       rateLimit,
	})

	// Set OnRequest callback
	collector.OnRequest(func(r *colly.Request) {
		log.Printf("Requesting URL: %s\n", r.URL.String()) // Log the request URL
	})

	return &Crawler{BaseURL: baseURL, Storage: storage, MaxDepth: maxDepth, RateLimit: rateLimit, Collector: collector}, nil
}

// Start method to begin crawling and indexing
func (c *Crawler) Start(url string) {
	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		content := e.Text

		hash := md5.Sum([]byte(url))
		docID := hex.EncodeToString(hash[:])

		err := c.Storage.IndexDocument("example_index", docID, map[string]interface{}{"url": url, "content": content})
		if err != nil {
			log.Println("Error indexing document:", err)
		}
	})

	err := c.Collector.Visit(url)
	if err != nil {
		log.Println("Error visiting URL:", err)
	}

	c.Collector.Wait()
}
