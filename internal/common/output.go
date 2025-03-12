// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This file specifically handles
// output formatting and user interaction through the command line interface.
package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// PrintErrorf prints an error message to stderr with formatting.
// This function should be used for displaying error messages to users
// in a consistent format across the application.
//
// Parameters:
//   - format: The format string for the error message
//   - args: Optional arguments for the format string
func PrintErrorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintWarnf prints a warning message in yellow.
// This function provides visual feedback for warning conditions
// by displaying the message in yellow text.
//
// Parameters:
//   - format: The format string for the warning message
//   - args: Optional arguments for the format string
func PrintWarnf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", text.FgYellow.Sprint("⚠"), fmt.Sprintf(format, args...))
}

// PrintSuccessf prints a success message with a green checkmark.
// This function provides visual feedback for successful operations
// by prefixing the message with a green checkmark symbol.
//
// Parameters:
//   - format: The format string for the success message
//   - args: Optional arguments for the format string
func PrintSuccessf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", text.FgGreen.Sprint("✓"), fmt.Sprintf(format, args...))
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

// PrintDivider prints a horizontal line divider to stdout.
// This function is used to visually separate sections of output
// for better readability in command-line interfaces.
//
// Parameters:
//   - width: The width of the divider in characters
func PrintDivider(width int) {
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", width))
}

// PrintTableHeaderf prints a formatted table header to stdout.
// This function is used to create consistent table headers for
// tabular data output in command-line interfaces.
//
// Parameters:
//   - format: The format string for the table header
//   - args: Optional arguments for the format string
func PrintTableHeaderf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
