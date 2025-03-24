// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This file specifically handles
// output formatting and user interaction through the command line interface.
package common

import (
	"fmt"
	"os"
	"strings"
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

// PrintInfof prints an informational message to stdout with formatting.
// This function should be used for displaying general information and
// status updates to users in a consistent format.
//
// Parameters:
//   - format: The format string for the info message
//   - args: Optional arguments for the format string
func PrintInfof(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// PrintConfirmation prints a confirmation prompt and returns the response.
// This function is used to get user confirmation for important operations,
// returning true for 'y' and false for any other input.
//
// Parameters:
//   - message: The confirmation message to display
//
// Returns:
//   - bool: true if the user confirms (enters 'y'), false otherwise
func PrintConfirmation(message string) bool {
	_, _ = fmt.Fprintf(os.Stdout, "%s (y/N): ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	return strings.ToLower(response) == "y"
}
