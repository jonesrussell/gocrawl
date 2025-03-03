package collector

import (
	"strings"

	"github.com/gocolly/colly/v2"
)

const (
	maxDepthError          = "Max depth limit reached"
	forbiddenDomainError   = "Forbidden domain"
	urlAlreadyVisitedError = "URL already visited"
	articleFoundKey        = "articleFound"
	bodyElementKey         = "bodyElement"
	htmlElementKey         = "htmlElement"
)

// configureContentProcessing sets up content processing for the collector
func configureContentProcessing(c *colly.Collector, p Params) {
	ignoredErrors := map[string]bool{
		maxDepthError:          true,
		forbiddenDomainError:   true,
		urlAlreadyVisitedError: true,
	}

	setupLinkFollowing(c, p, ignoredErrors)
	setupHTMLProcessing(c, p)
	setupArticleProcessing(c, p)
}

// setupLinkFollowing sets up link following logic for the collector
func setupLinkFollowing(c *colly.Collector, p Params, ignoredErrors map[string]bool) {
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		p.Logger.Debug("Link found", "text", e.Text, "link", link)
		visitErr := e.Request.Visit(e.Request.AbsoluteURL(link))
		if visitErr != nil && !ignoredErrors[visitErr.Error()] {
			p.Logger.Error("Failed to visit link", "link", link, "error", visitErr)
		}
	})
}

// setupHTMLProcessing sets up HTML element processing logic for the collector
func setupHTMLProcessing(c *colly.Collector, p Params) {
	c.OnHTML("html", func(e *colly.HTMLElement) {
		p.Logger.Debug("Found HTML element", "url", e.Request.URL.String())
		e.Request.Ctx.Put(htmlElementKey, e)
	})
}

// setupArticleProcessing sets up article processing logic for the collector
func setupArticleProcessing(c *colly.Collector, p Params) {
	articleSelector := getSelector(p.Source.Selectors.Article, DefaultArticleSelector)
	p.Logger.Debug("Using article selector", "selector", articleSelector)

	c.OnHTML(articleSelector, func(e *colly.HTMLElement) {
		if !isArticleMatched(e, articleSelector) {
			p.Logger.Debug("Article selector did not match", "url", e.Request.URL.String(), "selector", articleSelector)
			return
		}

		p.Logger.Debug("Found article", "url", e.Request.URL.String(), "selector", articleSelector)
		e.Request.Ctx.Put(articleFoundKey, "true")
	})

	c.OnScraped(func(r *colly.Response) {
		if p.ArticleProcessor == nil && p.ContentProcessor == nil {
			p.Logger.Debug("Skipping processing - no processors available", "url", r.Request.URL.String())
			return
		}

		e, ok := r.Ctx.GetAny(htmlElementKey).(*colly.HTMLElement)
		if !ok || e == nil {
			p.Logger.Debug("No HTML element found for processing", "url", r.Request.URL.String())
			return
		}

		switch {
		case r.Ctx.Get(articleFoundKey) == "true" && p.ArticleProcessor != nil:
			p.Logger.Debug("Processing as article", "url", r.Request.URL.String(), "title", e.ChildText("title"))
			if htmlEl, ok := r.Ctx.GetAny(htmlElementKey).(*colly.HTMLElement); ok && htmlEl != nil {
				p.ArticleProcessor.Process(htmlEl)
			} else {
				p.ArticleProcessor.Process(e)
			}

		case p.ContentProcessor != nil:
			p.Logger.Debug("Processing as content", "url", r.Request.URL.String(), "title", e.ChildText("title"))
			p.ContentProcessor.Process(e)

		default:
			p.Logger.Debug("No suitable processor found", "url", r.Request.URL.String(),
				"is_article", r.Ctx.Get(articleFoundKey),
				"has_article_processor", p.ArticleProcessor != nil,
				"has_content_processor", p.ContentProcessor != nil)
		}
	})
}

// isArticleMatched checks if the HTML element matches the article selector
func isArticleMatched(e *colly.HTMLElement, articleSelector string) bool {
	for _, selector := range strings.Split(articleSelector, ", ") {
		if e.DOM.Is(selector) {
			return true
		}
	}
	return false
}
