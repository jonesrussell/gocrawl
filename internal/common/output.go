package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// PrintError prints an error message to stderr
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintSuccess prints a success message with a green checkmark
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", text.FgGreen.Sprint("✓"), fmt.Sprintf(format, args...))
}

// PrintInfo prints an informational message
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// PrintConfirmation prints a confirmation prompt and returns the response
func PrintConfirmation(message string) bool {
	fmt.Print(message + " (y/N): ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

// PrintDivider prints a divider line
func PrintDivider(width int) {
	fmt.Println(strings.Repeat("-", width))
}

// PrintTable prints a table header
func PrintTableHeader(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
