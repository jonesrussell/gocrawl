package common

import "time"

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second

	// DefaultStartupTimeout is the default timeout for startup operations
	DefaultStartupTimeout = 30 * time.Second

	// DefaultOperationTimeout is the default timeout for general operations
	DefaultOperationTimeout = 5 * time.Second
)
