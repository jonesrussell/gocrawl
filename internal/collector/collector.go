package collector

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config"
)

// New creates a new collector instance
func New(p Params) (Result, error) {
	if err := ValidateParams(p); err != nil {
		return Result{}, err
	}

	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return Result{}, fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", p.BaseURL)
	}

	// Create collector with base configuration
	c := createBaseCollector(p, parsedURL.Hostname())

	// Configure collector settings
	if err := configureCollector(c, p); err != nil {
		return Result{}, err
	}

	// Create completion channel and handlers
	done := make(chan struct{})
	configureHandlers(c, p, done)

	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	return Result{Collector: c, Done: done}, nil
}

func createBaseCollector(p Params, domain string) *colly.Collector {
	return colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(p.MaxDepth),
		colly.AllowedDomains(domain),
	)
}

func configureCollector(c *colly.Collector, p Params) error {
	// Configure rate limiting
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RandomDelay,
		Parallelism: p.Parallelism,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Configure debugger and logging
	if p.Debugger != nil {
		c.SetDebugger(p.Debugger)
	}
	ConfigureLogging(c, p.Logger)

	// Parse rate limit duration
	rateLimit, err := time.ParseDuration(p.Source.RateLimit)
	if err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	// Convert sources.Config to config.Source
	source := config.Source{
		Name:         p.Source.Name,
		URL:          p.Source.URL,
		ArticleIndex: p.Source.ArticleIndex,
		Index:        p.Source.Index,
		RateLimit:    rateLimit,
		MaxDepth:     p.Source.MaxDepth,
		Time:         p.Source.Time,
		Selectors: config.SourceSelectors{
			Article: config.ArticleSelectors{
				Container:     p.Source.Selectors.Article.Container,
				Title:         p.Source.Selectors.Article.Title,
				Body:          p.Source.Selectors.Article.Body,
				Intro:         p.Source.Selectors.Article.Intro,
				Byline:        p.Source.Selectors.Article.Byline,
				PublishedTime: p.Source.Selectors.Article.PublishedTime,
				TimeAgo:       p.Source.Selectors.Article.TimeAgo,
				JSONLD:        p.Source.Selectors.Article.JSONLd,
				Section:       p.Source.Selectors.Article.Section,
				Keywords:      p.Source.Selectors.Article.Keywords,
				Description:   p.Source.Selectors.Article.Description,
				OGTitle:       p.Source.Selectors.Article.OgTitle,
				OGDescription: p.Source.Selectors.Article.OgDescription,
				OGImage:       p.Source.Selectors.Article.OgImage,
				OgURL:         p.Source.Selectors.Article.OgURL,
				Canonical:     p.Source.Selectors.Article.Canonical,
				WordCount:     p.Source.Selectors.Article.WordCount,
				PublishDate:   p.Source.Selectors.Article.PublishDate,
				Category:      p.Source.Selectors.Article.Category,
				Tags:          p.Source.Selectors.Article.Tags,
				Author:        p.Source.Selectors.Article.Author,
				BylineName:    p.Source.Selectors.Article.BylineName,
			},
		},
	}

	// Configure content processing
	contentParams := ContentParams{
		Logger:           p.Logger,
		Source:           source,
		ArticleProcessor: p.ArticleProcessor,
		ContentProcessor: p.ContentProcessor,
	}
	configureContentProcessing(c, contentParams)

	return nil
}

func configureHandlers(c *colly.Collector, p Params, done chan struct{}) {
	// Add completion handler to ensure proper completion signaling
	c.OnScraped(func(r *colly.Response) {
		// Check if this is the last request
		if r.Request.URL.String() == p.BaseURL {
			p.Logger.Debug("Base URL scraped, crawl complete")
			// Signal completion by closing the done channel
			close(done)
		}
	})

	// Add error handler to ensure we know about any failures
	c.OnError(func(r *colly.Response, err error) {
		p.Logger.Error("Request failed", "url", r.Request.URL, "error", err)
	})

	// Add request handler to track progress
	c.OnRequest(func(r *colly.Request) {
		p.Logger.Debug("Starting request", "url", r.URL)
	})

	// Add response handler to track completion
	c.OnResponse(func(r *colly.Response) {
		p.Logger.Debug("Received response", "url", r.Request.URL, "status", r.StatusCode)
	})
}
