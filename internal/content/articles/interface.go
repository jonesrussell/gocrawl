// Package articles provides functionality for processing and managing article content.
package articles

import "github.com/gocolly/colly/v2"

// Interface defines the interface for processing articles.
// It handles the extraction and processing of article content from web pages.
type Interface interface {
	// Process handles the processing of article content.
	// It takes a colly.HTMLElement and processes the article found within it.
	Process(e *colly.HTMLElement) error
}
