// Package testutils provides testing utilities for the sources package.
package testutils

import "github.com/jonesrussell/gocrawl/internal/sources"

// NewTestSources creates a new Sources instance for testing.
// This function is intended for testing purposes only.
func NewTestSources(configs []sources.Config) *sources.Sources {
	s := &sources.Sources{}
	s.SetSources(configs)
	return s
}
