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
func isArticleType(e *colly.HTMLElement, p ContentParams) bool {
	// Check meta type
	metaType := e.ChildAttr(`meta[property="og:type"]`, "content")
	if metaType == "article" {
		return true
	}

	// Check schema.org type
	schemaType := e.ChildAttr(`meta[name="type"]`, "content")
	if schemaType == "NewsArticle" || schemaType == "Article" {
		return true
	}

	// Check meta name type
	if e.ChildAttr(`meta[name="type"]`, "content") == "article" {
		return true
	}

	// Check for Village Media specific article class
	if e.DOM.Find(".details").Length() > 0 && e.DOM.Find(".details-title").Length() > 0 {
		return true
	}

	// Check if URL contains typical article patterns
	url := e.Request.URL.String()

	// Check for listing page patterns first - if found, definitely not an article
	listingPatterns := []string{
		"/category/",
		"/tag/",
		"/topics/",
		"/search/",
		"/archive/",
		"/author/",
		"/index/",
		"/feed/",
		"/rss/",
	}

	for _, pattern := range listingPatterns {
		if strings.Contains(url, pattern) {
			return false
		}
	}

	// Only check article patterns if we haven't matched a listing pattern
	articlePatterns := []string{
		"/article/",
		"/news/",
		"/story/",
		"/opp-beat/", // Include police beat articles
		"/local-news/",
	}

	// Additional validation for URL patterns
	for _, pattern := range articlePatterns {
		if strings.Contains(url, pattern) {
			// Additional checks to ensure it's really an article:
			// 1. Check if there's a date element
			if e.DOM.Find("time").Length() > 0 {
				return true
			}
			// 2. Check for article body using source-specific selector or default
			bodySelector := DefaultBodySelector
			if p.Source.Selectors.Article.Body != "" {
				bodySelector = p.Source.Selectors.Article.Body
			}
			if e.DOM.Find(bodySelector).Length() > 0 {
				return true
			}
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
	setupArticleProcessing(c, p, log)
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

// setupArticleProcessing sets up article processing logic for the collector
func setupArticleProcessing(c *colly.Collector, p ContentParams, log *contentLogger) {
	articleSelector := getSelector(p.Source.Selectors.Article, DefaultArticleSelector)
	log.debug("Using article selector", "selector", articleSelector)

	setupArticleDetection(c, articleSelector, log, p)
	setupContentProcessing(c, p, log)
}

// setupArticleDetection sets up article processing logic for the collector
func setupArticleDetection(c *colly.Collector, articleSelector string, log *contentLogger, p ContentParams) {
	// Set up detection for the entire HTML document
	c.OnHTML("html", func(e *colly.HTMLElement) {
		// First check metadata and URL patterns
		if !isArticleType(e, p) {
			log.debug("Content type is not an article", "url", e.Request.URL.String())
			return
		}

		// If a specific article selector is provided, use it as additional validation
		if articleSelector != "" && !isArticleMatched(e, articleSelector) {
			log.debug("Article selector did not match",
				"url", e.Request.URL.String(),
				"selector", articleSelector,
				"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
				"schema_type", e.ChildAttr(`meta[name="type"]`, "content"),
			)
			return
		}

		log.debug("Found article",
			"url", e.Request.URL.String(),
			"selector", articleSelector,
			"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
			"schema_type", e.ChildAttr(`meta[name="type"]`, "content"),
		)
		cm := newContextManager(e.Request.Ctx)
		cm.markAsArticle()
	})
}

func setupContentProcessing(c *colly.Collector, p ContentParams, log *contentLogger) {
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

func canProcess(p ContentParams) bool {
	return p.ArticleProcessor != nil || p.ContentProcessor != nil
}

func processContent(e *colly.HTMLElement, r *colly.Response, p ContentParams, cm *contextManager, log *contentLogger) {
	// Early return if no processors are available
	if p.ArticleProcessor == nil && p.ContentProcessor == nil {
		log.debug("No processors available", "url", r.Request.URL.String())
		return
	}

	switch {
	case cm.isArticle():
		log.debug("Processing as article", "url", r.Request.URL.String(), "title", e.ChildText("title"))
		p.ArticleProcessor.Process(e)
		return // Exit after processing as article

	case !cm.isArticle(): // Only process as content if NOT marked as article
		log.debug("Processing as content", "url", r.Request.URL.String(), "title", e.ChildText("title"))
		p.ContentProcessor.Process(e)
		return // Exit after processing as content

	default:
		log.debug("No suitable processor found", "url", r.Request.URL.String(),
			"is_article", cm.isArticle())
	}
}

// isArticleMatched checks if the HTML element matches the article selector
func isArticleMatched(e *colly.HTMLElement, articleSelector string) bool {
	for _, selector := range strings.Split(articleSelector, ", ") {
		if e.DOM.Is(selector) {
			return true
		}
	}
	return false
}
