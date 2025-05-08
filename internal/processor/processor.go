// Package processor provides processor interfaces used across the application.
package processor

import (
	"context"
	"io"

	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
)

// Processor defines the interface for content processors.
type Processor interface {
	// Process processes the given content.
	Process(ctx context.Context, content any) error
	// CanProcess returns whether the processor can handle the given content type.
	CanProcess(contentType contenttype.Type) bool
	// ContentType returns the content type this processor handles.
	ContentType() contenttype.Type
	// Start initializes the processor.
	Start(ctx context.Context) error
	// Stop stops the processor.
	Stop(ctx context.Context) error
	// ValidateJob validates a job before processing.
	ValidateJob(job *content.Job) error
	// ParseHTML parses HTML content.
	ParseHTML(r io.Reader) error
	// ExtractLinks extracts links from the content.
	ExtractLinks() ([]string, error)
	// ExtractContent extracts the main content.
	ExtractContent() (string, error)
	// RegisterProcessor registers a new processor.
	RegisterProcessor(processor content.ContentProcessor)
	// GetProcessor returns a processor for the given content type.
	GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error)
}

// ContentProcessor defines the interface for processing content.
type ContentProcessor interface {
	// Process processes the given content.
	Process(ctx context.Context, content string) error
	// CanProcess returns whether the processor can handle the given content type.
	CanProcess(contentType contenttype.Type) bool
	// ContentType returns the content type this processor handles.
	ContentType() contenttype.Type
	// ValidateJob validates a job before processing.
	ValidateJob(job *content.Job) error
}
