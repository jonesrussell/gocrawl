package models

import "github.com/gocolly/colly/v2"

// ContentProcessor processes content and articles
type ContentProcessor interface {
	ProcessContent(e *colly.HTMLElement)
	ProcessArticle(e *colly.HTMLElement)
}
