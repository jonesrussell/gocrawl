package collector

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/fx"
)

// DebuggerInterface is an interface for the debugger
type DebuggerInterface interface {
	Init() error
	OnRequest(e *colly.Request)
	OnResponse(e *colly.Response)
	OnError(e *colly.Response, err error)
	OnEvent(e *debug.Event)
	Event(e *debug.Event)
}

// ContentProcessor processes non-article content
type ContentProcessor interface {
	ProcessContent(e *colly.HTMLElement)
}

const (
	// Default selectors when none are specified in the source config
	DefaultArticleSelector    = "body"
	DefaultTitleSelector      = "h1"
	DefaultDateSelector       = "time"
	DefaultAuthorSelector     = "span.author"
	DefaultCategoriesSelector = "div.categories"

	// Context keys
	articleFoundKey = "articleFound"
	bodyElementKey  = "bodyElement"
)

// getSelector returns the specified selector or falls back to a default
func getSelector(specified, defaultSelector string) string {
	if specified == "" {
		return defaultSelector
	}
	return specified
}

// Params holds the parameters for creating a Collector
type Params struct {
	fx.In

	ArticleProcessor *article.Processor
	ContentProcessor ContentProcessor
	BaseURL          string
	Context          context.Context
	Debugger         *logger.CollyDebugger
	Logger           logger.Interface
	MaxDepth         int
	Parallelism      int
	RandomDelay      time.Duration
	RateLimit        time.Duration
	Source           *sources.Config
}

// Result holds the collector instance
type Result struct {
	fx.Out

	Collector *colly.Collector
}

// New creates a new collector instance
func New(p Params) (Result, error) {
	// Validate URL
	if p.BaseURL == "" {
		return Result{}, errors.New("base URL cannot be empty")
	}

	// Check if ArticleProcessor is nil
	if p.ArticleProcessor == nil {
		return Result{}, errors.New("article processor is required")
	}

	// Check if Logger is nil
	if p.Logger == nil {
		return Result{}, errors.New("logger is required")
	}

	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return Result{}, fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", p.BaseURL)
	}

	// Extract the domain from the BaseURL
	domain := parsedURL.Hostname()

	// Create collector with base configuration
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(p.MaxDepth),
		colly.AllowedDomains(domain), // Set the allowed domain
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RandomDelay,
		Parallelism: p.Parallelism,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to set rate limit: %w", err)
	}

	if p.Debugger != nil {
		c.SetDebugger(p.Debugger)
	}

	// Configure logging
	ConfigureLogging(c, p.Logger)

	// Set up link following
	var ignoredErrors = map[string]bool{
		"Max depth limit reached": true,
		"Forbidden domain":        true,
		"URL already visited":     true,
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		p.Logger.Debug("Link found", "text", e.Text, "link", link)
		visitErr := e.Request.Visit(e.Request.AbsoluteURL(link))
		if visitErr != nil && !ignoredErrors[visitErr.Error()] {
			p.Logger.Error("Failed to visit link", "link", link, "error", visitErr)
		}
	})

	// Get the article selector with fallback
	articleSelector := getSelector(p.Source.Selectors.Article, DefaultArticleSelector)

	// Store body element when found
	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(bodyElementKey, e)
	})

	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(articleFoundKey, "true")
		p.Logger.Debug("Found article", "url", e.Request.URL.String(), "selector", articleSelector)
		p.ArticleProcessor.ProcessPage(e)
	})

	// Handle non-article content after response is received
	c.OnResponse(func(r *colly.Response) {
		if r.Ctx.Get(articleFoundKey) != "true" && p.ContentProcessor != nil {
			p.Logger.Debug("Processing non-article content", "url", r.Request.URL.String())
			// Create a temporary collector to parse the response
			temp := colly.NewCollector()
			temp.OnHTML("body", func(e *colly.HTMLElement) {
				p.ContentProcessor.ProcessContent(e)
			})
			err := temp.Visit("data:text/html;charset=utf-8," + string(r.Body))
			if err != nil {
				p.Logger.Error("Failed to process non-article content", "url", r.Request.URL.String(), "error", err)
			}
		}
	})

	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	return Result{Collector: c}, nil
}

// ConfigureLogging sets up logging for the collector
func ConfigureLogging(c *colly.Collector, log logger.Interface) {
	c.OnRequest(func(r *colly.Request) {
		log.Debug("Requesting URL", "url", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Debug("Received response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Error("Error occurred",
			"url", r.Request.URL.String(),
			"error", err.Error(),
		)
	})
}
