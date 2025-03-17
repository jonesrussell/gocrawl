// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// Constants for content processing and error handling
const (
	// Error messages for various crawling scenarios
	maxDepthError          = "Max depth limit reached"
	forbiddenDomainError   = "Forbidden domain"
	urlAlreadyVisitedError = "URL already visited"

	// Keys for storing data in colly context
	articleFoundKey = "articleFound"
	htmlElementKey  = "htmlElement"

	// Logging tag for content-related operations
	logTag = "collector/content"

	// Default CSS selector for finding article content
	DefaultBodySelector = "article, .article"
)

// URL patterns used for content classification
var (
	// Patterns that indicate a page is a listing/index page
	listingPatterns = []string{
		"/category/", "/tag/", "/topics/", "/search/",
		"/archive/", "/author/", "/index/", "/feed/", "/rss/",
	}

	// Patterns that indicate a page is an article
	articlePatterns = []string{
		"/article/", "/news/", "/story/", "/opp-beat/", "/local-news/",
	}
)

// ContentParams contains the parameters for configuring content collection.
// It holds the necessary dependencies and processors for handling different types of content.
type ContentParams struct {
	// Logger for content-related operations
	Logger logger.Interface
	// Source configuration for the collector
	Source config.Source
	// Processor for handling article content
	ArticleProcessor models.ContentProcessor
	// Processor for handling general content
	ContentProcessor models.ContentProcessor
}

// contextManager handles storing and retrieving data from colly context.
// It provides a clean interface for managing context data during the crawling process.
type contextManager struct {
	// ctx holds the colly context instance
	ctx *colly.Context
}

// newContextManager creates a new context manager instance.
//
// Parameters:
//   - ctx: The colly context to manage
//
// Returns:
//   - *contextManager: The initialized context manager
func newContextManager(ctx *colly.Context) *contextManager {
	return &contextManager{ctx: ctx}
}

// setHTMLElement stores an HTML element in the context.
//
// Parameters:
//   - e: The HTML element to store
func (cm *contextManager) setHTMLElement(e *colly.HTMLElement) {
	cm.ctx.Put(htmlElementKey, e)
}

// getHTMLElement retrieves the stored HTML element from the context.
//
// Returns:
//   - *colly.HTMLElement: The stored HTML element
//   - bool: Whether the element was found and is valid
func (cm *contextManager) getHTMLElement() (*colly.HTMLElement, bool) {
	e, ok := cm.ctx.GetAny(htmlElementKey).(*colly.HTMLElement)
	return e, ok && e != nil
}

// markAsArticle marks the current content as an article in the context.
func (cm *contextManager) markAsArticle() {
	cm.ctx.Put(articleFoundKey, "true")
}

// isArticle checks if the current content is marked as an article.
//
// Returns:
//   - bool: Whether the content is marked as an article
func (cm *contextManager) isArticle() bool {
	return cm.ctx.Get(articleFoundKey) == "true"
}

// isArticleType checks if the content appears to be an article based on metadata.
// It performs multiple checks in order of reliability:
// 1. Checks OpenGraph meta type
// 2. Checks schema.org type
// 3. Checks URL patterns
// 4. Validates with additional content markers
//
// Parameters:
//   - e: The HTML element to analyze
//
// Returns:
//   - bool: Whether the content appears to be an article
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

// contentLogger wraps logging functionality with consistent tags.
// It provides a standardized way to log content-related operations.
type contentLogger struct {
	// p holds the content parameters including the logger
	p ContentParams
}

// newLogger creates a new content logger instance.
//
// Parameters:
//   - p: The content parameters containing the logger
//
// Returns:
//   - *contentLogger: The initialized logger
func newLogger(p ContentParams) *contentLogger {
	return &contentLogger{p: p}
}

// debug logs a debug message with the content tag.
//
// Parameters:
//   - msg: The message to log
//   - keysAndValues: Additional key-value pairs to include in the log
func (l *contentLogger) debug(msg string, keysAndValues ...any) {
	args := append([]any{"tag", logTag}, keysAndValues...)
	l.p.Logger.Debug(msg, args...)
}

// error logs an error message with the content tag.
//
// Parameters:
//   - msg: The message to log
//   - keysAndValues: Additional key-value pairs to include in the log
func (l *contentLogger) error(msg string, keysAndValues ...any) {
	args := append([]any{"tag", logTag}, keysAndValues...)
	l.p.Logger.Error(msg, args...)
}

// setupArticleProcessing sets up article processing logic for the collector.
// It configures the collector to detect and process article content.
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
func setupArticleProcessing(c *colly.Collector, log *contentLogger) {
	log.debug("Setting up article processing")
	setupArticleDetection(c, log)
}

// processScrapedContent handles the processing of scraped content.
// It determines the content type and routes it to the appropriate processor.
//
// Parameters:
//   - c: The collector to configure
//   - p: The content parameters
//   - log: The logger for content operations
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

// configureContentProcessing sets up content processing for the collector.
// It configures all necessary handlers for content processing, including:
// - Link following
// - HTML processing
// - Article detection
// - Content processing
//
// Parameters:
//   - c: The collector to configure
//   - p: The content parameters
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

// setupLinkFollowing sets up link following logic for the collector.
// It configures the collector to follow links while handling various error cases.
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
//   - ignoredErrors: Map of errors that should be ignored during link following
func setupLinkFollowing(c *colly.Collector, log *contentLogger, ignoredErrors map[string]bool) {
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.debug("Link found", "text", e.Text, "link", link)

		if err := visitLink(e, link, ignoredErrors); err != nil {
			log.error("Failed to visit link", "link", link, "error", err)
		}
	})
}

// visitLink attempts to visit a link while handling ignored errors.
//
// Parameters:
//   - e: The HTML element containing the link
//   - link: The link URL to visit
//   - ignoredErrors: Map of errors that should be ignored
//
// Returns:
//   - error: Any non-ignored error that occurred during the visit
func visitLink(e *colly.HTMLElement, link string, ignoredErrors map[string]bool) error {
	if err := e.Request.Visit(e.Request.AbsoluteURL(link)); err != nil {
		if !ignoredErrors[err.Error()] {
			return err
		}
	}
	return nil
}

// setupHTMLProcessing sets up HTML element processing logic for the collector.
// It stores the HTML element in the context for later processing.
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
func setupHTMLProcessing(c *colly.Collector, log *contentLogger) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		log.debug("Found HTML element", "url", e.Request.URL.String())
		cm := newContextManager(e.Request.Ctx)
		cm.setHTMLElement(e)
	})
}

// setupArticleDetection sets up article detection logic for the collector.
// It analyzes the content to determine if it's an article and marks it accordingly.
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
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

// canProcess checks if there are any content processors available.
//
// Parameters:
//   - p: The content parameters
//
// Returns:
//   - bool: Whether there are processors available
func canProcess(p ContentParams) bool {
	return p.ArticleProcessor != nil || p.ContentProcessor != nil
}

// processContent handles the processing of HTML content based on its type.
// It routes the content to the appropriate processor based on whether it's an article.
//
// Parameters:
//   - e: The HTML element to process
//   - r: The response containing the content
//   - p: The content parameters
//   - cm: The context manager
//   - log: The logger for content operations
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
