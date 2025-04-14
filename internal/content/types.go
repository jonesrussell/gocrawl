// Package content provides content processing types and interfaces.
package content

import (
	"github.com/jonesrussell/gocrawl/internal/common/contenttype"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
)

// Type aliases for commonly used interfaces and types.
// These aliases provide a convenient way to reference core types
// throughout the application while maintaining clear dependencies.
type (
	// Job is an alias for jobtypes.Job, representing a crawling job.
	Job = jobtypes.Job

	// Item is an alias for jobtypes.Item, representing a crawled item.
	Item = jobtypes.Item
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

// ContentType is an alias for contenttype.Type.
type ContentType = contenttype.Type

// Content type constants.
const (
	ContentTypeArticle = contenttype.Article
	ContentTypePage    = contenttype.Page
	ContentTypeVideo   = contenttype.Video
	ContentTypeImage   = contenttype.Image
	ContentTypeHTML    = contenttype.HTML
	ContentTypeJob     = contenttype.Job
)
