// Package content provides functionality for processing and managing web content.
package content

import "github.com/gocolly/colly/v2"

// Processor defines the interface for processing web content.
// It handles the extraction and processing of content from web pages.
type Processor interface {
	// Process handles the processing of web content.
	// It takes a colly.HTMLElement and processes the content found within it.
	Process(e *colly.HTMLElement)
}
