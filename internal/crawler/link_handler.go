// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"errors"
	"net/url"
	"strings"
	"time"

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

	// Validate URL if configured
	if h.crawler.cfg.ValidateURLs {
		if _, err := url.Parse(absLink); err != nil {
			h.crawler.logger.Debug("Invalid URL",
				"url", absLink,
				"error", err)
			return
		}
	}

	// Try to visit the URL with retries
	var lastErr error
	for i := range h.crawler.cfg.MaxRetries {
		err := e.Request.Visit(absLink)
		if err == nil {
			h.crawler.logger.Debug("Successfully visited link",
				"url", absLink)
			return
		}

		// Check if error is non-retryable by checking both error type and message
		// Colly may return errors with different message formats
		errMsg := err.Error()
		isNonRetryable := errors.Is(err, ErrAlreadyVisited) ||
			errors.Is(err, ErrMaxDepth) ||
			errors.Is(err, ErrMissingURL) ||
			errors.Is(err, ErrForbiddenDomain) ||
			strings.Contains(errMsg, "forbidden domain") ||
			strings.Contains(errMsg, "Forbidden domain") ||
			strings.Contains(errMsg, "max depth") ||
			strings.Contains(errMsg, "Max depth") ||
			strings.Contains(errMsg, "already visited") ||
			strings.Contains(errMsg, "Already visited")

		if isNonRetryable {
			// These are expected conditions, log at debug level
			h.crawler.logger.Debug("Skipping non-retryable link",
				"url", absLink,
				"error", errMsg)
			return
		}

		lastErr = err
		h.crawler.logger.Debug("Failed to visit link, retrying",
			"url", absLink,
			"error", err,
			"attempt", i+1,
			"max_retries", h.crawler.cfg.MaxRetries)

		// Wait before retrying
		time.Sleep(h.crawler.cfg.RetryDelay)
	}

	// If we get here, all retries failed
	h.crawler.logger.Error("Failed to visit link after retries",
		"url", absLink,
		"error", lastErr,
		"max_retries", h.crawler.cfg.MaxRetries)
}
