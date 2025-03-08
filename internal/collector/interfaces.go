// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

// ArticleProcessor defines the interface for processing articles during crawling.
// It handles the extraction and processing of article content from web pages.
type ArticleProcessor interface {
	// Process handles the processing of an article.
	// It takes an article interface{} which can be any type representing an article,
	// and returns an error if processing fails.
	Process(article interface{}) error
}

// ContentProcessor defines the interface for processing general web content.
// It handles the extraction and processing of content from web pages.
type ContentProcessor interface {
	// Process handles the processing of web content.
	// It takes a content string and returns the processed content and any error
	// that occurred during processing.
	Process(content string) (string, error)
}
