// Package sources provides the sources command implementation.
package sources

import (
	cmdvalidate "github.com/jonesrussell/gocrawl/cmd/validate"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates a new validate subcommand for sources.
func NewValidateCommand() *cobra.Command {
	cmd := cmdvalidate.ValidateCmd
	// Update the command name and usage
	cmd.Use = "validate"
	cmd.Short = "Validate CSS selectors against real articles"
	cmd.Long = `Tests CSS selectors from a source configuration against real article URLs
to verify they work correctly.

Example:
  # Validate selectors for a source (fetches sample articles from source URL)
  gocrawl sources validate --source "Mid-North Monitor" --samples 5

  # Validate selectors against specific URLs
  gocrawl sources validate --source "Mid-North Monitor" --urls "https://example.com/article1" "https://example.com/article2"`
	return cmd
}
