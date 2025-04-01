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
	"errors"
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

// ContextManager handles storing and retrieving data from colly context.
// It provides a clean interface for managing context data during the crawling process.
// The context manager is used to:
// - Store HTML elements for processing
// - Track article status
// - Maintain state between different callbacks
type ContextManager struct {
	// ctx holds the colly context instance
	ctx *colly.Context
}

// NewContextManager creates a new context manager instance.
// It initializes the context manager with the provided colly context.
//
// Parameters:
//   - ctx: The colly context to manage
//
// Returns:
//   - *ContextManager: The initialized context manager
func NewContextManager(ctx *colly.Context) *ContextManager {
	return &ContextManager{ctx: ctx}
}

// SetHTMLElement stores an HTML element in the context.
// This is used to make the HTML element available to other callbacks
// that need to process the content.
//
// Parameters:
//   - e: The HTML element to store
func (cm *ContextManager) SetHTMLElement(e *colly.HTMLElement) {
	cm.ctx.Put(htmlElementKey, e)
}

// GetHTMLElement retrieves the stored HTML element from the context.
// It ensures the element exists and is valid before returning it.
//
// Returns:
//   - *colly.HTMLElement: The stored HTML element
//   - bool: Whether the element was found and is valid
func (cm *ContextManager) GetHTMLElement() (*colly.HTMLElement, bool) {
	e, ok := cm.ctx.GetAny(htmlElementKey).(*colly.HTMLElement)
	return e, ok && e != nil
}

// MarkAsArticle marks the current content as an article in the context.
// This is used to indicate that the current page contains article content
// and should be processed accordingly.
func (cm *ContextManager) MarkAsArticle() {
	cm.ctx.Put(articleFoundKey, "true")
}

// IsArticle checks if the current content is marked as an article.
// This is used to determine how the content should be processed.
//
// Returns:
//   - bool: Whether the content is marked as an article
func (cm *ContextManager) IsArticle() bool {
	return cm.ctx.Get(articleFoundKey) == "true"
}

// IsArticleType checks if the content appears to be an article based on metadata.
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
func IsArticleType(e *colly.HTMLElement) bool {
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

// ContentLogger wraps logging functionality with consistent tags.
// It provides a standardized way to log content-related operations
// with consistent formatting and context.
type ContentLogger struct {
	// p holds the content parameters including the logger
	p ContentParams
}

// NewLogger creates a new content logger instance.
// It initializes the logger with the provided content parameters.
//
// Parameters:
//   - p: The content parameters containing the logger
//
// Returns:
//   - *ContentLogger: The initialized logger
func NewLogger(p ContentParams) *ContentLogger {
	return &ContentLogger{p: p}
}

// Debug logs a debug message with the content tag.
// It automatically adds the content tag to all log messages
// for consistent logging across the package.
//
// Parameters:
//   - msg: The message to log
//   - args: Additional arguments for the log message
func (l *ContentLogger) Debug(msg string, args ...any) {
	l.p.Logger.Debug(msg, append(args, "tag", logTag)...)
}

// Error logs an error message with the content tag.
// It automatically adds the content tag to all log messages
// for consistent logging across the package.
//
// Parameters:
//   - msg: The message to log
//   - args: Additional arguments for the log message
func (l *ContentLogger) Error(msg string, args ...any) {
	l.p.Logger.Error(msg, append(args, "tag", logTag)...)
}

// VisitLink attempts to visit a link from an HTML element.
// The function:
// 1. Resolves the absolute URL
// 2. Visits the URL
// 3. Filters out ignored errors
//
// Parameters:
//   - e: The HTML element containing the link
//   - link: The link URL to visit
//   - ignoredErrors: Map of errors that should be ignored
//
// Returns:
//   - error: Any non-ignored error that occurred during the visit
func VisitLink(e *colly.HTMLElement, link string, ignoredErrors map[string]bool) error {
	absURL := e.Request.AbsoluteURL(link)
	collector, ok := e.Response.Ctx.GetAny("collector").(*colly.Collector)
	if !ok {
		return errors.New("failed to get collector from context")
	}
	if err := collector.Visit(absURL); err != nil {
		if !ignoredErrors[err.Error()] {
			return err
		}
	}
	return nil
}

// CanProcess checks if there are any content processors available.
// This is used to determine if content processing should be attempted.
//
// Parameters:
//   - p: The content parameters
//
// Returns:
//   - bool: Whether there are processors available
func CanProcess(p ContentParams) bool {
	return p.ArticleProcessor != nil || p.ContentProcessor != nil
}

// ProcessContent handles the processing of HTML content based on its type.
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
func ProcessContent(e *colly.HTMLElement, r *colly.Response, p ContentParams, cm *ContextManager, log *ContentLogger) {
	url := r.Request.URL.String()
	title := e.ChildText("title")

	// Double check if it's an article using IsArticleType
	isArticle := cm.IsArticle() && IsArticleType(e)

	if isArticle {
		if p.ArticleProcessor == nil {
			log.Debug("No article processor available", "url", url)
			return
		}
		log.Debug("Processing as article", "url", url, "title", title)
		if err := p.ArticleProcessor.Process(e); err != nil {
			log.Error("Failed to process article", "url", url, "error", err)
			return
		}
		return
	}

	if p.ContentProcessor == nil {
		log.Debug("No content processor available", "url", url)
		return
	}
	log.Debug("Processing as content", "url", url, "title", title)
	if err := p.ContentProcessor.Process(e); err != nil {
		log.Error("Failed to process content", "url", url, "error", err)
		return
	}
}

// setupArticleProcessing configures article processing for the collector.
// It:
// 1. Sets up article detection
// 2. Configures content processing
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
func setupArticleProcessing(c *colly.Collector, log *ContentLogger) {
	log.Debug("Setting up article processing")
	setupArticleDetection(c, log)
}

// processScrapedContent configures content processing for the collector.
// It:
// 1. Checks if processors are available
// 2. Retrieves the HTML element from context
// 3. Routes content to appropriate processor
//
// Parameters:
//   - c: The collector to configure
//   - p: The content parameters
//   - log: The logger for content operations
func processScrapedContent(c *colly.Collector, p ContentParams, log *ContentLogger) {
	c.OnScraped(func(r *colly.Response) {
		if !CanProcess(p) {
			log.Debug("Skipping processing - no processors available", "url", r.Request.URL.String())
			return
		}

		cm := NewContextManager(r.Request.Ctx)
		e, ok := cm.GetHTMLElement()
		if !ok {
			log.Debug("No HTML element found for processing", "url", r.Request.URL.String())
			return
		}

		ProcessContent(e, r, p, cm, log)
	})
}

// setupArticleDetection configures article detection for the collector.
// It:
// 1. Checks for article metadata
// 2. Marks content as article if detected
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
func setupArticleDetection(c *colly.Collector, log *ContentLogger) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		cm := NewContextManager(e.Request.Ctx)

		// Check if it's an article type
		if IsArticleType(e) {
			log.Debug("Found article",
				"url", e.Request.URL.String(),
				"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
				"schema_type", e.ChildAttr(`meta[name="type"]`, "content"))
			cm.MarkAsArticle()
			return
		}

		log.Debug("Content type is not an article",
			"url", e.Request.URL.String(),
			"meta_type", e.ChildAttr(`meta[property="og:type"]`, "content"),
			"schema_type", e.ChildAttr(`meta[name="type"]`, "content"))
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

	log := NewLogger(p)

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
func setupLinkFollowing(c *colly.Collector, log *ContentLogger, ignoredErrors map[string]bool) {
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Debug("Link found", "text", e.Text, "link", link)

		if err := VisitLink(e, link, ignoredErrors); err != nil {
			log.Error("Failed to visit link", "link", link, "error", err)
		}
	})
}

// setupHTMLProcessing configures HTML element processing for the collector.
// It:
// 1. Finds HTML elements
// 2. Stores them in the context for later processing
//
// Parameters:
//   - c: The collector to configure
//   - log: The logger for content operations
func setupHTMLProcessing(c *colly.Collector, log *ContentLogger) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		log.Debug("Found HTML element", "url", e.Request.URL.String())
		cm := NewContextManager(e.Request.Ctx)
		cm.SetHTMLElement(e)
	})
}
