// Package sources provides the sources command implementation.
package sources

import (
	cmdgenerate "github.com/jonesrussell/gocrawl/cmd/generate"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates a new generate subcommand for sources.
func NewGenerateCommand() *cobra.Command {
	cmd := cmdgenerate.GenerateCmd
	// Update the command name and usage
	cmd.Use = "generate"
	cmd.Short = "Generate CSS selectors for a new source"
	cmd.Long = `Analyzes a news source and generates initial CSS selectors.

Example:
  # Write to file for review
  gocrawl sources generate https://www.example.com/news -o new_source.yaml

  # Analyze both listing and article pages for best results
  gocrawl sources generate https://www.example.com/news \
    --article-url https://www.example.com/news/article-123 \
    -o new_source.yaml

  # Append directly to sources.yaml (with confirmation and backup)
  gocrawl sources generate https://www.example.com/news \
    --article-url https://www.example.com/news/article-123 \
    --append`
	return cmd
}
