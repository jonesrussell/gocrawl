// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"errors"
	"strings"

	"github.com/gocolly/colly/v2"
)

// LinkHandler handles link processing for the crawler.
type LinkHandler struct {
	crawler *Crawler
}

// NewLinkHandler creates a new link handler.
func NewLinkHandler(c *Crawler) *LinkHandler {
	return &LinkHandler{
		crawler: c,
	}
}

// HandleLink processes a single link from an HTML element.
func (h *LinkHandler) HandleLink(e *colly.HTMLElement) {
	link := e.Attr("href")
	if link == "" {
		return
	}

	// Skip anchor links and javascript
	if strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
		return
	}

	// Skip mailto and tel links
	if strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "tel:") {
		return
	}

	// Make absolute URL if relative
	absLink := e.Request.AbsoluteURL(link)
	if absLink == "" {
		h.crawler.logger.Debug("Failed to make absolute URL",
			"url", link)
		return
	}

	err := e.Request.Visit(absLink)
	if err != nil {
		if errors.Is(err, colly.ErrAlreadyVisited) ||
			errors.Is(err, colly.ErrMaxDepth) ||
			errors.Is(err, colly.ErrMissingURL) ||
			errors.Is(err, colly.ErrForbiddenDomain) {
			return
		}

		h.crawler.logger.Error("Failed to visit link",
			"url", absLink,
			"error", err)
	} else {
		h.crawler.logger.Debug("Successfully visited link",
			"url", absLink)
	}
}
