package collector

import (
	"strings"

	"github.com/gocolly/colly/v2"
)

// configureContentProcessing sets up content processing for the collector
func configureContentProcessing(c *colly.Collector, p Params) {
	// Set up link following
	var ignoredErrors = map[string]bool{
		"Max depth limit reached": true,
		"Forbidden domain":        true,
		"URL already visited":     true,
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		p.Logger.Debug("Link found", "text", e.Text, "link", link)
		visitErr := e.Request.Visit(e.Request.AbsoluteURL(link))
		if visitErr != nil && !ignoredErrors[visitErr.Error()] {
			p.Logger.Error("Failed to visit link", "link", link, "error", visitErr)
		}
	})

	// Get the article selector with fallback
	articleSelector := getSelector(p.Source.Selectors.Article, DefaultArticleSelector)
	p.Logger.Debug("Using article selector", "selector", articleSelector)

	// Store HTML element when found
	c.OnHTML("html", func(e *colly.HTMLElement) {
		p.Logger.Debug("Found HTML element", "url", e.Request.URL.String())
		e.Request.Ctx.Put(bodyElementKey, e)
	})

	// Mark when we find an article
	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		p.Logger.Debug("Checking article selector match",
			"url", e.Request.URL.String(),
			"selector", articleSelector)

		// Check if the element itself matches the selector
		matchedSelector := ""
		for _, selector := range strings.Split(articleSelector, ", ") {
			if e.DOM.Is(selector) {
				matchedSelector = selector
				break
			}
		}

		if matchedSelector == "" {
			p.Logger.Debug("Article selector did not match",
				"url", e.Request.URL.String(),
				"selector", articleSelector)
			return
		}

		p.Logger.Debug("Found article",
			"url", e.Request.URL.String(),
			"selector", articleSelector,
			"matched_selector", matchedSelector)

		e.Request.Ctx.Put(articleFoundKey, "true")

		// Get the full HTML element for metadata
		if htmlEl, ok := e.Request.Ctx.GetAny(bodyElementKey).(*colly.HTMLElement); ok && htmlEl != nil {
			p.ArticleProcessor.ProcessArticle(htmlEl)
		} else {
			p.ArticleProcessor.ProcessArticle(e)
		}
	})

	// Final decision point - process as content if not already processed as article
	c.OnScraped(func(r *colly.Response) {
		// Skip if no content processor or if already processed as article
		if p.ContentProcessor == nil || r.Ctx.Get(articleFoundKey) == "true" {
			return
		}

		// Get the stored HTML element
		if e, ok := r.Ctx.GetAny(bodyElementKey).(*colly.HTMLElement); ok && e != nil {
			p.Logger.Debug("Processing as content",
				"url", r.Request.URL.String(),
				"title", e.ChildText("title"))
			p.ContentProcessor.ProcessContent(e)
		}
	})
}
