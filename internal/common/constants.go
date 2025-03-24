// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This package serves as a central
// location for common types, interfaces, and helper functions.
package common

import "time"

// Timeout constants define the default durations for various operations
// throughout the application. These values can be overridden through
// configuration if needed.
const (
	// DefaultOperationTimeout is the default timeout for general operations.
	// This duration is used for common operations like API calls,
	// data processing tasks, or crawler shutdown that should complete
	// in a reasonable time.
	DefaultOperationTimeout = 30 * time.Second
)
