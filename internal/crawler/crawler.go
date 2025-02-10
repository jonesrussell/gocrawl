package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
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
	BaseURL     string
	Storage     storage.Storage
	MaxDepth    int
	RateLimit   time.Duration
	Collector   *colly.Collector
	Logger      logger.Interface
	IndexName   string
	articleChan chan *models.Article
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

// NewCrawler creates a new Crawler instance
func NewCrawler(p Params) (Result, error) {
	if p.Logger == nil {
		return Result{}, errors.New("logger is required")
	}

	// Create a new collector
	c := colly.NewCollector(
		colly.MaxDepth(p.MaxDepth),
		colly.Async(true),
	)

	// Set rate limiting
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RateLimit,
	})
	if err != nil {
		return Result{}, fmt.Errorf("error setting rate limit: %w", err)
	}

	crawler := &Crawler{
		BaseURL:     p.BaseURL,
		Storage:     p.Storage,
		MaxDepth:    p.MaxDepth,
		RateLimit:   p.RateLimit,
		Collector:   c,
		Logger:      p.Logger,
		IndexName:   p.Config.IndexName,
		articleChan: make(chan *models.Article, 100),
	}

	// Configure collector callbacks
	c.OnRequest(func(r *colly.Request) {
		crawler.Logger.Debug("Requesting URL", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.Logger.Debug("Received response", r.Request.URL.String(), r.StatusCode)
		if r.StatusCode != HTTPStatusOK {
			crawler.Logger.Warn(
				"Non-200 response received",
				r.Request.URL.String(),
				r.StatusCode,
			)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error occurred", r.Request.URL.String(), err)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}

		crawler.Logger.Debug("Found link", "url", link)

		if err := e.Request.Visit(link); err != nil {
			if !errors.Is(err, colly.ErrAlreadyVisited) &&
				!errors.Is(err, colly.ErrMissingURL) &&
				!errors.Is(err, colly.ErrForbiddenDomain) {
				crawler.Logger.Debug("Could not visit link", "url", link, "error", err.Error())
			}
		}
	})

	if p.Debugger != nil {
		c.SetDebugger(p.Debugger)
	}

	p.Logger.Info("Crawler initialized",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit)

	return Result{Crawler: crawler}, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, shutdowner fx.Shutdowner) error {
	// Check storage connection
	if err := c.Storage.TestConnection(ctx); err != nil {
		c.Logger.Error("Failed to connect to storage", "error", err)
		return err
	}

	// Check if index exists
	exists, err := c.Storage.IndexExists(ctx, c.IndexName)
	if err != nil {
		c.Logger.Error("Failed to check index existence", "error", err)
		return err
	}
	if !exists {
		c.Logger.Error("Index does not exist",
			"index", c.IndexName,
			"message", "Please create the index before running the crawler")
		return fmt.Errorf("index %s does not exist", c.IndexName)
	}

	c.Logger.Info("Starting crawl with valid index", "index", c.IndexName)

	c.articleChan = make(chan *models.Article, 100)

	// Create a done channel for article processor
	processorDone := make(chan struct{})

	// Start article processor
	go func() {
		c.processArticles(ctx)
		close(processorDone)
	}()

	// Configure collector
	c.Collector.OnHTML("article", c.processPage)
	c.configureCollectors(ctx)

	c.Logger.Info("Starting crawling process")

	// Create error channel for async crawling
	errChan := make(chan error, 1)
	crawlerDone := make(chan bool, 1)

	go func() {
		if err := c.Collector.Visit(c.BaseURL); err != nil {
			errChan <- err
			return
		}
		c.Collector.Wait()
		crawlerDone <- true
	}()

	// Wait for either completion or context cancellation
	var result error
	select {
	case err := <-errChan:
		result = err
	case <-crawlerDone:
		c.Logger.Info("Crawling completed successfully")
	case <-ctx.Done():
		c.Logger.Error("Context cancelled", "error", ctx.Err())
		result = ctx.Err()
	}

	// Close article channel and wait for processor to finish
	close(c.articleChan)
	<-processorDone

	if result != nil {
		return result
	}
	return shutdowner.Shutdown()
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

	c.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return // Skip empty or invalid URLs
		}

		c.Logger.Debug("Found link", "url", link)

		// Visit the link asynchronously
		if err := e.Request.Visit(link); err != nil {
			if !errors.Is(err, colly.ErrAlreadyVisited) &&
				!errors.Is(err, colly.ErrMissingURL) &&
				!errors.Is(err, colly.ErrForbiddenDomain) {
				c.Logger.Debug("Could not visit link", "url", link, "error", err.Error())
			}
		}
	})

	c.Collector.OnHTML("html", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			c.Logger.Warn("Crawling stopped due to context cancellation", "error", ctx.Err())
			return
		}

		content := e.Text
		docID := generateDocumentID(e.Request.URL.String())

		if len(content) == 0 {
			c.Logger.Warn("Content is empty, skipping indexing", "url", e.Request.URL.String())
			return
		}

		c.Logger.Debug("Indexing document", "url", e.Request.URL.String(), "id", docID)
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

// processPage handles article extraction
func (c *Crawler) processPage(e *colly.HTMLElement) {
	// Common selectors for news articles
	article := &models.Article{
		ID:            uuid.New().String(),
		Title:         e.ChildText("h1.article-title, h1.headline, h1"),      // Common title selectors
		Body:          e.ChildText("div.article-body, div.content, article"), // Common content selectors
		Source:        e.Request.URL.String(),
		PublishedDate: extractDate(e),
		Author:        extractAuthor(e),
		Tags:          extractTags(e),
	}

	// Skip if no content found
	if article.Title == "" || article.Body == "" {
		return
	}

	c.Logger.Debug("Found article",
		"title", article.Title,
		"url", article.Source)

	c.articleChan <- article
}

// Helper functions for extraction
func extractDate(e *colly.HTMLElement) time.Time {
	// Common date selectors
	dateStr := e.ChildAttr("meta[property='article:published_time']", "content")
	if dateStr == "" {
		dateStr = e.ChildText("time, .date, .published-date")
	}

	// Try parsing with different formats
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"January 2, 2006",
	} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}

func extractAuthor(e *colly.HTMLElement) string {
	// Common author selectors
	author := e.ChildAttr("meta[name='author']", "content")
	if author == "" {
		author = e.ChildText(".author-name, .byline, .author")
	}
	return author
}

func extractTags(e *colly.HTMLElement) []string {
	var tags []string

	// Check meta tags
	e.ForEach("meta[property='article:tag'], meta[name='keywords']", func(_ int, el *colly.HTMLElement) {
		content := el.Attr("content")
		if content != "" {
			// Split if comma-separated
			for _, tag := range strings.Split(content, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	})

	// Check tag links
	e.ForEach("a.tag, .tags a", func(_ int, el *colly.HTMLElement) {
		tag := strings.TrimSpace(el.Text)
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	return tags
}

// processArticles handles the bulk indexing of articles
func (c *Crawler) processArticles(ctx context.Context) {
	var articles []*models.Article
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Helper function to bulk index articles
	indexBatch := func() {
		if len(articles) > 0 {
			if err := c.Storage.BulkIndexArticles(ctx, articles); err != nil {
				c.Logger.Error("Failed to bulk index articles", "error", err)
			} else {
				c.Logger.Info("Successfully indexed articles", "count", len(articles))
			}
			articles = articles[:0] // Clear the slice while keeping capacity
		}
	}

	for {
		select {
		case <-ctx.Done():
			indexBatch() // Final index attempt before exit
			return
		case article, ok := <-c.articleChan:
			if !ok {
				indexBatch() // Final index attempt before exit
				return
			}
			articles = append(articles, article)
			if len(articles) >= 10 {
				indexBatch()
			}
		case <-ticker.C:
			indexBatch()
		}
	}
}
