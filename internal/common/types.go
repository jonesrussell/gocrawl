package common

import (
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/config"
	loggertypes "github.com/jonesrussell/gocrawl/internal/logger/types"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Type aliases for commonly used interfaces and types.
// These aliases provide a convenient way to reference core types
// throughout the application while maintaining clear dependencies.
type (
	// Config is an alias for the configuration interface, providing
	// access to application-wide settings.
	Config = config.Interface

	// Logger is an alias for logger/types.Logger, providing structured logging capabilities.
	Logger = loggertypes.Logger

	// Storage is an alias for storage.Interface, providing
	// data persistence operations across the application.
	Storage = storagetypes.Interface

	// Job is an alias for jobtypes.Job, representing a crawling job.
	Job = jobtypes.Job

	// Item is an alias for jobtypes.Item, representing a crawled item.
	Item = jobtypes.Item
)
