package common

import (
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
