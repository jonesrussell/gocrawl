package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crypto/tls"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
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
	articleSvc  article.Service
	indexSvc    *storage.IndexService
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

	// Parse domain from BaseURL
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return Result{}, fmt.Errorf("invalid base URL: %w", err)
	}
	domain := parsedURL.Host

	// Create a new collector with proper configuration
	c := colly.NewCollector(
		colly.MaxDepth(p.MaxDepth),
		colly.Async(true),
		colly.AllowedDomains(domain),
		colly.MaxBodySize(10*1024*1024),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
		colly.AllowURLRevisit(),
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RateLimit,
		Parallelism: 2,
	})
	if err != nil {
		return Result{}, fmt.Errorf("error setting rate limit: %w", err)
	}

	// Add transport configuration
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	articleSvc := article.NewService(p.Logger)
	indexSvc := storage.NewIndexService(p.Storage, p.Logger)

	crawler := &Crawler{
		BaseURL:     p.BaseURL,
		Storage:     p.Storage,
		MaxDepth:    p.MaxDepth,
		RateLimit:   p.RateLimit,
		Collector:   c,
		Logger:      p.Logger,
		IndexName:   p.Config.Crawler.IndexName,
		articleChan: make(chan *models.Article, 100),
		articleSvc:  articleSvc,
		indexSvc:    indexSvc,
	}

	// Configure collector callbacks
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		crawler.Logger.Debug("Visiting", "url", r.URL.String(), "headers", r.Headers)
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.Logger.Debug("Got response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"headers", r.Headers,
			"body_length", len(r.Body))

		// Log first 500 chars of response to see what we're getting
		if len(r.Body) > 0 {
			preview := string(r.Body)
			if len(preview) > 500 {
				preview = preview[:500]
			}
			crawler.Logger.Debug("Response preview", "body", preview)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error scraping", "url", r.Request.URL.String(), "error", err)
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
		"rateLimit", p.RateLimit,
		"domain", domain)

	return Result{Crawler: crawler}, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context, shutdowner fx.Shutdowner) error {
	c.Logger.Debug("Testing storage connection")
	if err := c.Storage.TestConnection(ctx); err != nil {
		return fmt.Errorf("storage connection failed: %w", err)
	}

	c.Logger.Debug("Ensuring index exists")
	if err := c.indexSvc.EnsureIndex(ctx, c.IndexName); err != nil {
		return fmt.Errorf("index setup failed: %w", err)
	}

	c.Logger.Info("Starting crawl with valid index", "index", c.IndexName)

	c.articleChan = make(chan *models.Article, 100)
	c.Logger.Debug("Created article channel")

	// Create a done channel for article processor
	processorDone := make(chan struct{})

	// Start article processor
	go func() {
		c.Logger.Debug("Starting article processor")
		c.processArticles(ctx)
		c.Logger.Debug("Article processor finished")
		close(processorDone)
	}()

	c.Logger.Info("Starting crawling process", "url", c.BaseURL)

	// Create error channel for async crawling
	errChan := make(chan error, 1)
	crawlerDone := make(chan bool, 1)

	go func() {
		c.Logger.Debug("Starting Visit", "url", c.BaseURL)
		if err := c.Collector.Visit(c.BaseURL); err != nil {
			c.Logger.Error("Visit failed", "error", err, "url", c.BaseURL)
			if errors.Is(err, colly.ErrMissingURL) {
				c.Logger.Error("Missing URL error")
			} else if errors.Is(err, colly.ErrForbiddenDomain) {
				c.Logger.Error("Forbidden domain error")
			}
			errChan <- err
			return
		}
		c.Logger.Debug("Waiting for collector to finish")
		c.Collector.Wait()
		c.Logger.Debug("Collector finished")
		crawlerDone <- true
	}()

	// Wait for either completion or context cancellation
	var result error
	select {
	case err := <-errChan:
		c.Logger.Error("Crawler error", "error", err)
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
	c.Logger.Debug("Processing page", "url", e.Request.URL.String())
	article := c.articleSvc.ExtractArticle(e)
	if article == nil {
		c.Logger.Debug("No article extracted", "url", e.Request.URL.String())
		return
	}
	c.Logger.Debug("Article extracted", "url", e.Request.URL.String(), "title", article.Title)
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

// Add these methods to the Crawler struct
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.Collector = collector
}

func (c *Crawler) SetArticleService(svc article.Service) {
	c.articleSvc = svc
}

func (c *Crawler) SetBaseURL(url string) {
	c.BaseURL = url
}
