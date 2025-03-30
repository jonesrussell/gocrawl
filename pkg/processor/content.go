// Package processor provides content processing functionality for the application.
package processor

import (
	"time"
)

// Content represents processed content from a source.
type Content struct {
	// Source is the source of the content.
	Source string
	// Title is the title of the content.
	Title string
	// Body is the main content body.
	Body string
	// URL is the URL of the content.
	URL string
	// PublishedAt is when the content was published.
	PublishedAt time.Time
	// Author is the author of the content.
	Author string
	// Categories are the content categories.
	Categories []string
	// Tags are the content tags.
	Tags []string
	// Metadata contains additional content metadata.
	Metadata map[string]string
}

// ProcessedContent represents the result of processing content.
type ProcessedContent struct {
	// Content is the processed content.
	Content Content
	// ProcessingTime is when the content was processed.
	ProcessingTime time.Time
	// ProcessingDuration is how long processing took.
	ProcessingDuration time.Duration
	// Error is any error that occurred during processing.
	Error error
}

// Processor defines the interface for content processing.
type Processor interface {
	// Process processes the given content.
	Process(content []byte) (*ProcessedContent, error)
}
