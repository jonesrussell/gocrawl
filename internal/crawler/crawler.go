package crawler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
	articleSvc  *article.Service
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

	articleSvc := article.NewService(p.Logger)
	indexSvc := storage.NewIndexService(p.Storage, p.Logger)

	crawler := &Crawler{
		BaseURL:     p.BaseURL,
		Storage:     p.Storage,
		MaxDepth:    p.MaxDepth,
		RateLimit:   p.RateLimit,
		Collector:   c,
		Logger:      p.Logger,
		IndexName:   p.Config.IndexName,
		articleChan: make(chan *models.Article, 100),
		articleSvc:  articleSvc,
		indexSvc:    indexSvc,
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
	if err := c.Storage.TestConnection(ctx); err != nil {
		return fmt.Errorf("storage connection failed: %w", err)
	}

	if err := c.indexSvc.EnsureIndex(ctx, c.IndexName); err != nil {
		return fmt.Errorf("index setup failed: %w", err)
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
	article := c.articleSvc.ExtractArticle(e)
	if article == nil {
		return
	}
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
