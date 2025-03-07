package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// PrintErrorf prints an error message to stderr
func PrintErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintSuccessf prints a success message with a green checkmark
func PrintSuccessf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", text.FgGreen.Sprint("✓"), fmt.Sprintf(format, args...))
}

// PrintInfof prints an informational message
func PrintInfof(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// PrintConfirmation prints a confirmation prompt and returns the response
func PrintConfirmation(message string) bool {
	_, _ = fmt.Fprintf(os.Stdout, "%s (y/N): ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	return strings.ToLower(response) == "y"
}

// PrintDivider prints a divider line
func PrintDivider(width int) {
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", width))
}

// PrintTableHeaderf prints a table header
func PrintTableHeaderf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
