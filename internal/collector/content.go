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

	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		e.Request.Ctx.Put(articleFoundKey, "true")
		p.Logger.Debug("Found article", "url", e.Request.URL.String(), "selector", articleSelector)
		p.ArticleProcessor.ProcessPage(e)
	})

	// Handle non-article content after response is received
	c.OnResponse(func(r *colly.Response) {
		if r.Ctx.Get(articleFoundKey) != "true" && p.ContentProcessor != nil {
			p.Logger.Debug("Processing non-article content", "url", r.Request.URL.String())
			// Create a temporary collector to parse the response
			temp := colly.NewCollector()
			temp.OnHTML("body", func(e *colly.HTMLElement) {
				p.ContentProcessor.ProcessContent(e)
			})
			err := temp.Visit("data:text/html;charset=utf-8," + string(r.Body))
			if err != nil {
				p.Logger.Error("Failed to process non-article content", "url", r.Request.URL.String(), "error", err)
			}
		}
	})
}
