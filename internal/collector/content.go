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

	// Store HTML element when found
	c.OnHTML("html", func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(bodyElementKey, e)
	})

	// Mark when we find an article
	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(articleFoundKey, "true")
		// Get the specific selector that matched
		matchedSelector := ""
		for _, selector := range strings.Split(articleSelector, ", ") {
			// Check if the element itself matches the selector
			if e.DOM.Is(selector) {
				matchedSelector = selector
				break
			}
		}
		p.Logger.Debug("Found article",
			"url", e.Request.URL.String(),
			"selector", articleSelector,
			"matched_selector", matchedSelector)

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
			// Check which selectors were attempted but didn't match
			attemptedSelectors := strings.Split(articleSelector, ", ")
			matchedSelector := "html"
			for _, selector := range attemptedSelectors {
				// Check if the element itself matches the selector
				if e.DOM.Is(selector) {
					matchedSelector = selector
					break
				}
			}
			p.Logger.Debug("Found webpage",
				"url", r.Request.URL.String(),
				"selector", "html",
				"matched_selector", matchedSelector,
				"title", e.ChildText("title"),
				"h1", e.ChildText("h1"),
				"h2", e.ChildText("h2"),
				"h3", e.ChildText("h3"),
				"h4", e.ChildText("h4"),
				"h5", e.ChildText("h5"),
				"h6", e.ChildText("h6"),
			)
			p.ContentProcessor.ProcessContent(e)
		}
	})
}
