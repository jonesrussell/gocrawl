package crawler

import (
	"context"
	"encoding/json"
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
		colly.AllowedDomains("www.elliotlaketoday.com"),
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

	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		crawler.processPage(e)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}

		if strings.Contains(link, "/opp-beat") ||
			strings.Contains(link, "/police") ||
			strings.Contains(link, "/news") {
			crawler.Logger.Debug("Found article link", "url", link)
			if err := e.Request.Visit(link); err != nil {
				if !errors.Is(err, colly.ErrAlreadyVisited) {
					crawler.Logger.Debug("Could not visit article", "url", link, "error", err.Error())
				}
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

	// Create index if it doesn't exist
	if !exists {
		c.Logger.Info("Index does not exist, creating...", "index", c.IndexName)
		if err := c.createArticleIndex(ctx); err != nil {
			c.Logger.Error("Failed to create index", "error", err)
			return err
		}
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

// processPage handles article extraction
func (c *Crawler) processPage(e *colly.HTMLElement) {
	// Add debug logging at start
	c.Logger.Debug("Starting article processing",
		"url", e.Request.URL.String(),
		"has_title", e.DOM.Find("h1.details-title").Length() > 0,
		"has_body", e.DOM.Find("#details-body").Length() > 0)

	// Extract metadata from JSON-LD first
	var jsonLD struct {
		DateCreated   string   `json:"dateCreated"`
		DateModified  string   `json:"dateModified"`
		DatePublished string   `json:"datePublished"`
		Author        string   `json:"author"`
		Keywords      []string `json:"keywords"`
		Section       string   `json:"articleSection"`
	}

	e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			c.Logger.Debug("Failed to parse JSON-LD", "error", err)
		}
	})

	// Clean up author (remove date)
	author := e.ChildText(".details-byline")
	if idx := strings.Index(author, "    "); idx != -1 {
		author = strings.TrimSpace(author[:idx])
	}

	// Common selectors for news articles
	article := &models.Article{
		ID:     uuid.New().String(),
		Title:  e.ChildText("h1.details-title"),
		Body:   e.ChildText("#details-body"),
		Source: e.Request.URL.String(),
		Author: author,
		Tags:   make([]string, 0),
	}

	// Get intro/description
	if intro := e.ChildText(".details-intro"); intro != "" {
		article.Body = intro + "\n\n" + article.Body
	}

	// Add tags from multiple sources
	// 1. JSON-LD section
	if jsonLD.Section != "" {
		article.Tags = append(article.Tags, jsonLD.Section)
	}

	// 2. JSON-LD keywords
	if len(jsonLD.Keywords) > 0 {
		article.Tags = append(article.Tags, jsonLD.Keywords...)
	}

	// 3. Meta keywords
	if keywords := e.ChildAttr("meta[name='keywords']", "content"); keywords != "" {
		for _, tag := range strings.Split(keywords, "|") {
			if tag = strings.TrimSpace(tag); tag != "" {
				article.Tags = append(article.Tags, tag)
			}
		}
	}

	// 4. Breadcrumb navigation
	e.ForEach("ol.nav-breadcrumb li a", func(_ int, el *colly.HTMLElement) {
		if tag := strings.TrimSpace(el.Text); tag != "" && tag != "Home" {
			article.Tags = append(article.Tags, tag)
		}
	})

	// Remove duplicates from tags
	seen := make(map[string]bool)
	uniqueTags := make([]string, 0)
	for _, tag := range article.Tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	article.Tags = uniqueTags

	// Parse published date from time element
	if timeStr := e.ChildAttr("time.timeago", "datetime"); timeStr != "" {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			article.PublishedDate = t
		} else {
			c.Logger.Debug("Failed to parse datetime",
				"datetime", timeStr,
				"error", err)
		}
	} else if jsonLD.DatePublished != "" {
		if t, err := time.Parse(time.RFC3339, jsonLD.DatePublished); err == nil {
			article.PublishedDate = t
		}
	}

	// Skip empty articles
	if article.Title == "" && article.Body == "" {
		c.Logger.Debug("Skipping empty article", "url", article.Source)
		return
	}

	// Log article details at debug level
	c.Logger.Debug("Processing article",
		"id", article.ID,
		"title", article.Title,
		"url", article.Source,
		"date", article.PublishedDate,
		"author", article.Author,
		"tags", article.Tags)

	// Send article to channel for processing
	c.articleChan <- article
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

// createArticleIndex creates the articles index with proper mappings
func (c *Crawler) createArticleIndex(ctx context.Context) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"body": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"author": map[string]interface{}{
					"type": "keyword",
				},
				"published_date": map[string]interface{}{
					"type": "date",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	if err := c.Storage.CreateIndex(ctx, c.IndexName, mapping); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	c.Logger.Info("Created index", "index", c.IndexName)
	return nil
}
