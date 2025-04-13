// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
)

// HTMLProcessor handles HTML processing for the crawler.
type HTMLProcessor struct {
	crawler *Crawler
}

// NewHTMLProcessor creates a new HTML processor.
func NewHTMLProcessor(c *Crawler) *HTMLProcessor {
	return &HTMLProcessor{
		crawler: c,
	}
}

// ProcessHTML processes the HTML content.
func (p *HTMLProcessor) ProcessHTML(e *colly.HTMLElement) {
	// Detect content type and get appropriate processor
	processor := p.selectProcessor(e)
	if processor == nil {
		p.crawler.logger.Debug("No processor found for content",
			"url", e.Request.URL.String(),
			"type", p.detectContentType(e))
		p.crawler.state.IncrementProcessed()
		return
	}

	// Process the content
	err := processor.Process(p.crawler.state.Context(), e)
	if err != nil {
		p.crawler.logger.Error("Failed to process content",
			"error", err,
			"url", e.Request.URL.String(),
			"type", p.detectContentType(e))
		p.crawler.state.IncrementError()
	} else {
		p.crawler.logger.Debug("Successfully processed content",
			"url", e.Request.URL.String(),
			"type", p.detectContentType(e))
	}

	p.crawler.state.IncrementProcessed()
}

// selectProcessor selects the appropriate processor for the given HTML element
func (p *HTMLProcessor) selectProcessor(e *colly.HTMLElement) common.Processor {
	contentType := p.detectContentType(e)

	// Try to get a processor for the specific content type
	processor := p.getProcessorForType(contentType)
	if processor != nil {
		return processor
	}

	// Fallback: Try additional processors
	for _, proc := range p.crawler.processors {
		if proc.CanProcess(e) {
			return proc
		}
	}

	return nil
}

// getProcessorForType returns a processor for the given content type
func (p *HTMLProcessor) getProcessorForType(contentType common.ContentType) common.Processor {
	switch contentType {
	case common.ContentTypeArticle:
		return p.crawler.articleProcessor
	case common.ContentTypePage:
		return p.crawler.contentProcessor
	case common.ContentTypeVideo, common.ContentTypeImage, common.ContentTypeHTML, common.ContentTypeJob:
		// Try to find a processor for the specific content type
		for _, proc := range p.crawler.processors {
			if proc.ContentType() == contentType {
				return proc
			}
		}
	}
	return nil
}

// detectContentType detects the type of content in the HTML element
func (p *HTMLProcessor) detectContentType(e *colly.HTMLElement) common.ContentType {
	// First check HTTP headers
	contentType := e.Response.Headers.Get("Content-Type")
	if contentType != "" {
		switch {
		case strings.Contains(contentType, "text/html"):
			// Continue with HTML-specific checks
		case strings.Contains(contentType, "image/"):
			return common.ContentTypeImage
		case strings.Contains(contentType, "video/"):
			return common.ContentTypeVideo
		case strings.Contains(contentType, "application/pdf"):
			return common.ContentTypePage
		default:
			// If it's not HTML, we can't process it as HTML
			return common.ContentTypePage
		}
	}

	// Check for special pages in URL
	url := e.Request.URL.String()
	if strings.Contains(url, "/login") || strings.Contains(url, "/signin") || strings.Contains(url, "/register") {
		return common.ContentTypePage
	}

	// Check for article-specific elements and metadata
	hasArticleTag := e.DOM.Find("article").Length() > 0
	hasArticleClass := e.DOM.Find(".article").Length() > 0
	hasArticleMeta := e.DOM.Find("meta[property='og:type'][content='article']").Length() > 0
	hasPublicationDate := e.DOM.Find("time[datetime], .published-date, .post-date").Length() > 0
	hasAuthor := e.DOM.Find(".author, .byline, meta[name='author']").Length() > 0

	// If it has multiple article indicators, it's likely an article
	if (hasArticleTag || hasArticleClass) && (hasPublicationDate || hasAuthor || hasArticleMeta) {
		return common.ContentTypeArticle
	}

	// Check for video content - look for video players or embeds
	hasVideoPlayer := e.DOM.Find("video, .video-player, .video-container, iframe[src*='youtube'], iframe[src*='vimeo']").Length() > 0
	if hasVideoPlayer {
		return common.ContentTypeVideo
	}

	// Check for job listings - look for specific job-related elements
	hasJobElements := e.DOM.Find(".job-listing, .job-posting, .job-description, .job-title").Length() > 0
	if hasJobElements {
		return common.ContentTypeJob
	}

	// Check for image content - only if it's a dedicated image page
	// Don't classify as image just because there are images on the page
	hasImageGallery := e.DOM.Find(".image-gallery, .photo-gallery, .gallery-container").Length() > 0
	hasSingleImage := e.DOM.Find("img").Length() == 1 && e.DOM.Find("article, .article").Length() == 0
	if hasImageGallery || hasSingleImage {
		return common.ContentTypeImage
	}

	// Default to page content type
	return common.ContentTypePage
}
