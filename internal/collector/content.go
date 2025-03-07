package collector

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

const (
	maxDepthError          = "Max depth limit reached"
	forbiddenDomainError   = "Forbidden domain"
	urlAlreadyVisitedError = "URL already visited"
	articleFoundKey        = "articleFound"
	htmlElementKey         = "htmlElement"
	logTag                 = "collector/content"
	DefaultBodySelector    = "article, .article"
)

// URL patterns for content classification
var (
	listingPatterns = []string{
		"/category/", "/tag/", "/topics/", "/search/",
		"/archive/", "/author/", "/index/", "/feed/", "/rss/",
	}

	articlePatterns = []string{
		"/article/", "/news/", "/story/", "/opp-beat/", "/local-news/",
	}
)

// ContentParams contains the parameters for configuring content collection
type ContentParams struct {
	Logger           logger.Interface
	Source           config.Source
	ArticleProcessor models.ContentProcessor
	ContentProcessor models.ContentProcessor
}

// contextManager handles storing and retrieving data from colly context
type contextManager struct {
	ctx *colly.Context
}

func newContextManager(ctx *colly.Context) *contextManager {
	return &contextManager{ctx: ctx}
}

func (cm *contextManager) setHTMLElement(e *colly.HTMLElement) {
	cm.ctx.Put(htmlElementKey, e)
}

func (cm *contextManager) getHTMLElement() (*colly.HTMLElement, bool) {
	e, ok := cm.ctx.GetAny(htmlElementKey).(*colly.HTMLElement)
	return e, ok && e != nil
}

func (cm *contextManager) markAsArticle() {
	cm.ctx.Put(articleFoundKey, "true")
}

func (cm *contextManager) isArticle() bool {
	return cm.ctx.Get(articleFoundKey) == "true"
}

// isArticleType checks if the content appears to be an article based on metadata
func isArticleType(e *colly.HTMLElement) bool {
	// Check meta type first as it's most reliable
	if metaType := e.ChildAttr(`meta[property="og:type"]`, "content"); metaType == "article" {
		return true
	}

	// Check schema.org type
	schemaType := e.ChildAttr(`meta[name="type"]`, "content")
	if strings.Contains(strings.ToLower(schemaType), "article") {
		return true
	}

	// Check URL patterns - if it's a listing page, it's not an article
	url := e.Request.URL.String()

	// Early return if it's a listing page
	for _, pattern := range listingPatterns {
		if strings.Contains(url, pattern) {
			return false
		}
	}

	// Check for article patterns in URL
	for _, pattern := range articlePatterns {
		if strings.Contains(url, pattern) {
			// Additional validation to ensure it's an article
			hasTime := e.DOM.Find("time").Length() > 0
			hasDetails := e.DOM.Find(".details").Length() > 0
			return hasTime || hasDetails
		}
	}

	return false
}

// contentLogger wraps logging functionality with consistent tags
type contentLogger struct {
	p ContentParams
}

func newLogger(p ContentParams) *contentLogger {
	return &contentLogger{p: p}
}

func (l *contentLogger) debug(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"tag", logTag}, keysAndValues...)
	l.p.Logger.Debug(msg, args...)
}

func (l *contentLogger) error(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"tag", logTag}, keysAndValues...)
	l.p.Logger.Error(msg, args...)
}

// setupArticleProcessing sets up article processing logic for the collector
func setupArticleProcessing(c *colly.Collector, log *contentLogger) {
	log.debug("Setting up article processing")
	setupArticleDetection(c, log)
}

// processScrapedContent handles the processing of scraped content
func processScrapedContent(c *colly.Collector, p ContentParams, log *contentLogger) {
	c.OnScraped(func(r *colly.Response) {
		if !canProcess(p) {
			log.debug("Skipping processing - no processors available", "url", r.Request.URL.String())
			return
		}

		cm := newContextManager(r.Ctx)
		e, ok := cm.getHTMLElement()
		if !ok {
			log.debug("No HTML element found for processing", "url", r.Request.URL.String())
			return
		}

		processContent(e, r, p, cm, log)
	})
}

// configureContentProcessing sets up content processing for the collector
func configureContentProcessing(c *colly.Collector, p ContentParams) {
	ignoredErrors := map[string]bool{
		maxDepthError:          true,
		forbiddenDomainError:   true,
		urlAlreadyVisitedError: true,
	}

	log := newLogger(p)

	setupLinkFollowing(c, log, ignoredErrors)
	setupHTMLProcessing(c, log)
	setupArticleProcessing(c, log)
	processScrapedContent(c, p, log)
}

// setupLinkFollowing sets up link following logic for the collector
func setupLinkFollowing(c *colly.Collector, log *contentLogger, ignoredErrors map[string]bool) {
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.debug("Link found", "text", e.Text, "link", link)

		if err := visitLink(e, link, ignoredErrors); err != nil {
			log.error("Failed to visit link", "link", link, "error", err)
		}
	})
}

func visitLink(e *colly.HTMLElement, link string, ignoredErrors map[string]bool) error {
	if err := e.Request.Visit(e.Request.AbsoluteURL(link)); err != nil {
		if !ignoredErrors[err.Error()] {
			return err
		}
	}
	return nil
}

// setupHTMLProcessing sets up HTML element processing logic for the collector
func setupHTMLProcessing(c *colly.Collector, log *contentLogger) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		log.debug("Found HTML element", "url", e.Request.URL.String())
		cm := newContextManager(e.Request.Ctx)
		cm.setHTMLElement(e)
	})
}

// setupArticleDetection sets up article processing logic for the collector
func setupArticleDetection(c *colly.Collector, log *contentLogger) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		cm := newContextManager(e.Request.Ctx)

		// Check if it's an article type
		if isArticleType(e) {
			log.debug("Found article",
				"url", e.Request.URL.String(),
				"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
				"schema_type", e.ChildAttr(`meta[name="type"]`, "content"),
			)
			cm.markAsArticle()
			return
		}

		log.debug("Content type is not an article",
			"url", e.Request.URL.String(),
			"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
			"schema_type", e.ChildAttr(`meta[name="type"]`, "content"),
		)
	})
}

func canProcess(p ContentParams) bool {
	return p.ArticleProcessor != nil || p.ContentProcessor != nil
}

// processContent handles the processing of HTML content based on its type
func processContent(e *colly.HTMLElement, r *colly.Response, p ContentParams, cm *contextManager, log *contentLogger) {
	url := r.Request.URL.String()
	title := e.ChildText("title")

	// Double check if it's an article using isArticleType
	isArticle := cm.isArticle() && isArticleType(e)

	if isArticle {
		if p.ArticleProcessor == nil {
			log.debug("No article processor available", "url", url)
			return
		}
		log.debug("Processing as article", "url", url, "title", title)
		p.ArticleProcessor.Process(e)
		return
	}

	if p.ContentProcessor == nil {
		log.debug("No content processor available", "url", url)
		return
	}
	log.debug("Processing as content", "url", url, "title", title)
	p.ContentProcessor.Process(e)
}
