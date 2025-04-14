// Package contenttype provides content type definitions.
package contenttype

// Type represents the type of content being processed.
type Type string

const (
	// Article represents article content.
	Article Type = "article"
	// Page represents generic page content.
	Page Type = "page"
	// Video represents video content.
	Video Type = "video"
	// Image represents image content.
	Image Type = "image"
	// HTML represents HTML content.
	HTML Type = "html"
	// Job represents job content.
	Job Type = "job"
)
