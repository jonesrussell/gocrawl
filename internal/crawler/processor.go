// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

// ContentType represents the type of content being processed.
type ContentType string

const (
	// ContentTypeArticle represents article content.
	ContentTypeArticle ContentType = "article"
	// ContentTypeContent represents general content.
	ContentTypeContent ContentType = "content"
)

// processorMap maps content types to their processors.
type processorMap map[ContentType]Processor

// NewProcessorMap creates a new processor map.
func NewProcessorMap() processorMap {
	return make(processorMap)
}

// Add adds a processor for a specific content type.
func (m processorMap) Add(contentType ContentType, processor Processor) {
	m[contentType] = processor
}

// Get returns the processor for a specific content type.
func (m processorMap) Get(contentType ContentType) Processor {
	return m[contentType]
}

// Has returns true if a processor exists for the content type.
func (m processorMap) Has(contentType ContentType) bool {
	_, exists := m[contentType]
	return exists
}
