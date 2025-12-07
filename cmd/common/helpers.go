// Package common provides shared utilities for command implementations.
package common

// This file previously contained context.Value-based dependency injection helpers.
// Those helpers have been replaced with explicit dependency passing via CommandDeps.
//
// See cmd/common/deps.go for the new approach.
//
// Deprecated functions removed:
// - GetCommandContext() - Use CommandDeps directly instead
// - GetDependencies() - Use NewCommandDeps() from factory.go instead
