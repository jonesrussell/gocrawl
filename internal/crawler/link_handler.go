// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"errors"

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
	if link == "" || link == "#" {
		return
	}

	err := e.Request.Visit(link)
	if err != nil &&
		!errors.Is(err, colly.ErrAlreadyVisited) &&
		!errors.Is(err, colly.ErrMaxDepth) {
		h.crawler.logger.Error("Failed to visit link",
			"url", link,
			"error", err)
	}
}
