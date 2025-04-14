// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This file specifically handles
// output formatting and user interaction through the command line interface.
package common

import (
	"fmt"
	"os"
)

// PrintErrorf prints an error message to stderr with formatting.
// This function should be used for displaying error messages to users
// in a consistent format across the application.
//
// Parameters:
//   - format: The format string for the error message
//   - args: Optional arguments for the format string
func PrintErrorf(format string, args ...any) {
	_, err := fmt.Fprintf(os.Stderr, format+"\n", args...)
	if err != nil {
		return
	}
}
