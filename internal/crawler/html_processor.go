// Package crawler provides the core crawling functionality for the application.
package crawler

import (
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

	// Check for video content
	if e.DOM.Find("video").Length() > 0 || e.DOM.Find(".video").Length() > 0 {
		return common.ContentTypeVideo
	}

	// Check for image content
	if e.DOM.Find("img").Length() > 0 || e.DOM.Find(".image").Length() > 0 {
		return common.ContentTypeImage
	}

	// Check for job listings
	if e.DOM.Find(".job-listing").Length() > 0 || e.DOM.Find(".job-posting").Length() > 0 {
		return common.ContentTypeJob
	}

	// Default to page content type
	return common.ContentTypePage
}
