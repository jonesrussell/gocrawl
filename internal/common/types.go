package common

import (
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Type aliases for commonly used interfaces and types.
// These aliases provide a convenient way to reference core types
// throughout the application while maintaining clear dependencies.
type (
	// Config is an alias for the configuration interface, providing
	// access to application-wide settings.
	Config = config.Interface

	// Logger is an alias for the logger interface, providing
	// structured logging capabilities across the application.
	Logger = types.Logger

	// SourceManager is an alias for sources.Interface, providing
	// source configuration management across the application.
	SourceManager = sources.Interface

	// SourceConfig is an alias for sources.Config, providing
	// source configuration data structure across the application.
	SourceConfig = sources.Config

	// Storage is an alias for storage.Interface, providing
	// data persistence operations across the application.
	Storage = storagetypes.Interface
)
