// Package content provides content processing types and interfaces.
package content

import "github.com/jonesrussell/gocrawl/internal/common/jobtypes"

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
	JobType Type = "job"
)

// These types are defined as any to avoid import cycles.
// They will be used by other packages that need these types.
type (
	// Config represents the configuration interface.
	Config any

	// Storage represents the storage interface.
	Storage any
)

// JobValidator validates jobs before processing.
type JobValidator interface {
	// ValidateJob validates a job before processing.
	ValidateJob(job *jobtypes.Job) error
}
