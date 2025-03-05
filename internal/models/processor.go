package models

import "github.com/gocolly/colly/v2"

// ContentProcessor processes content
type ContentProcessor interface {
	Process(e *colly.HTMLElement)
}
