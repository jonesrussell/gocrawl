package crawler

import (
	"io"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Crawler struct to hold configuration or state if needed
type Crawler struct {
	BaseURL  string
	Storage  *storage.Storage
	MaxDepth int
}

// NewCrawler initializes a new Crawler
func NewCrawler(baseURL string, maxDepth int) (*Crawler, error) {
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	storage := storage.NewStorage(esClient)
	return &Crawler{BaseURL: baseURL, Storage: storage, MaxDepth: maxDepth}, nil
}

// Fetch fetches the content from the given URL
func (c *Crawler) Fetch(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return "", err
	}

	return string(body), nil
}

// Start method to begin crawling and indexing
func (c *Crawler) Start(url string) {
	err := c.Storage.IndexDocument("example_index", "1", map[string]interface{}{"url": url})
	if err != nil {
		log.Println("Error indexing document:", err)
	}
}
