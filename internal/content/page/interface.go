// Package page provides functionality for processing and managing web pages.
package page

import "github.com/gocolly/colly/v2"

// Interface defines the interface for processing web pages.
// It handles the extraction and processing of content from web pages.
type Interface interface {
	// Process handles the processing of web content.
	// It takes a colly.HTMLElement and processes the content found within it.
	Process(e *colly.HTMLElement)
}
