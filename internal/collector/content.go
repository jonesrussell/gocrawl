// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
//
// This file contains the content processing components that handle:
// - Article detection and classification
// - Content type routing
// - Link following and URL management
// - HTML element processing
package collector

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
)

// URL patterns used for content classification
var (
	// Patterns that indicate a page is a listing/index page.
	// These patterns are used to identify pages that contain lists of articles
	// rather than individual articles.
	listingPatterns = []string{
		"/category/", "/tag/", "/topics/", "/search/",
		"/archive/", "/author/", "/index/", "/feed/", "/rss/",
	}

	// Patterns that indicate a page is an article.
	// These patterns are used to identify pages that likely contain article content.
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
	ArticleProcessor Processor
	// Processor for handling general content
	ContentProcessor Processor
}

// contextManager handles storing and retrieving data from colly context.
// It provides a clean interface for managing context data during the crawling process.
// The context manager is used to:
// - Store HTML elements for processing
// - Track article status
// - Maintain state between different callbacks
type contextManager struct {
	// ctx holds the colly context instance
	ctx *colly.Context
}

// newContextManager creates a new context manager instance.
// It initializes the context manager with the provided colly context.
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
// This is used to make the HTML element available to other callbacks
// that need to process the content.
//
// Parameters:
//   - e: The HTML element to store
func (cm *contextManager) setHTMLElement(e *colly.HTMLElement) {
	cm.ctx.Put(htmlElementKey, e)
}

// getHTMLElement retrieves the stored HTML element from the context.
// It ensures the element exists and is valid before returning it.
//
// Returns:
//   - *colly.HTMLElement: The stored HTML element
//   - bool: Whether the element was found and is valid
func (cm *contextManager) getHTMLElement() (*colly.HTMLElement, bool) {
	e, ok := cm.ctx.GetAny(htmlElementKey).(*colly.HTMLElement)
	return e, ok && e != nil
}

// markAsArticle marks the current content as an article in the context.
// This is used to indicate that the current page contains article content
// and should be processed accordingly.
func (cm *contextManager) markAsArticle() {
	cm.ctx.Put(articleFoundKey, "true")
}

// isArticle checks if the current content is marked as an article.
// This is used to determine how the content should be processed.
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
// The function uses a combination of metadata and content markers
// to determine if a page contains article content. This helps ensure
// accurate content classification and processing.
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
// It provides a standardized way to log content-related operations
// with consistent formatting and context.
type contentLogger struct {
	// p holds the content parameters including the logger
	p ContentParams
}

// newLogger creates a new content logger instance.
// It initializes the logger with the provided content parameters.
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
// It automatically adds the content tag to all log messages
// for consistent logging across the package.
//
// Parameters:
//   - msg: The message to log
//   - keysAndValues: Additional key-value pairs to include in the log
func (l *contentLogger) debug(msg string, keysAndValues ...any) {
	args := append([]any{"tag", logTag}, keysAndValues...)
	l.p.Logger.Debug(msg, args...)
}

// error logs an error message with the content tag.
// It automatically adds the content tag to all log messages
// for consistent logging across the package.
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
// This includes setting up article detection and ensuring proper
// routing of article content to the appropriate processor.
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
// The function:
// 1. Checks if processors are available
// 2. Retrieves the HTML element from context
// 3. Determines content type (article vs general)
// 4. Routes to appropriate processor
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
// The function sets up a complete content processing pipeline that:
// 1. Follows links while handling errors
// 2. Processes HTML elements
// 3. Detects and processes articles
// 4. Handles general content
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
// The function:
// 1. Sets up HTML callback for link elements
// 2. Handles link processing
// 3. Manages error cases
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
// It:
// 1. Converts relative URLs to absolute URLs
// 2. Attempts to visit the link
// 3. Filters out ignored errors
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
// The function:
// 1. Sets up HTML callback for root element
// 2. Stores element in context
// 3. Logs processing status
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
// The function:
// 1. Sets up HTML callback for root element
// 2. Analyzes content type
// 3. Marks content as article if appropriate
// 4. Logs detection results
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
// This is used to determine if content processing should be attempted.
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
// The function:
// 1. Determines content type
// 2. Routes to appropriate processor
// 3. Handles processing errors
// 4. Logs processing status
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
		if err := p.ArticleProcessor.Process(e); err != nil {
			log.error("Failed to process article", "url", url, "error", err)
			return
		}
		return
	}

	if p.ContentProcessor == nil {
		log.debug("No content processor available", "url", url)
		return
	}
	log.debug("Processing as content", "url", url, "title", title)
	if err := p.ContentProcessor.Process(e); err != nil {
		log.error("Failed to process content", "url", url, "error", err)
		return
	}
}
