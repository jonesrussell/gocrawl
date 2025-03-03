package collector

import (
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

	// Store body element when found
	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(bodyElementKey, e)
	})

	// Mark when we find an article
	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(articleFoundKey, "true")
		p.Logger.Debug("Found article", "url", e.Request.URL.String(), "selector", articleSelector)
		p.ArticleProcessor.ProcessPage(e)
	})

	// Final decision point - process as content if not already processed as article
	c.OnScraped(func(r *colly.Response) {
		// Skip if no content processor or if already processed as article
		if p.ContentProcessor == nil || r.Ctx.Get(articleFoundKey) == "true" {
			return
		}

		// Get the stored body element
		if e, ok := r.Ctx.GetAny(bodyElementKey).(*colly.HTMLElement); ok && e != nil {
			p.Logger.Debug("Processing non-article content", "url", r.Request.URL.String())
			p.ContentProcessor.ProcessContent(e)
		}
	})
}
