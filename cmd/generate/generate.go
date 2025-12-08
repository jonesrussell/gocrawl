// Package generate provides the generate-selectors command implementation.
package generate

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jonesrussell/gocrawl/internal/generator"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	samples    int
)

// GenerateCmd represents the generate-selectors command.
var GenerateCmd = &cobra.Command{
	Use:   "generate-selectors [url]",
	Short: "Generate CSS selectors for a new source",
	Long: `Analyzes a news source and generates initial CSS selectors.

Example:
  # Write to file for review
  gocrawl generate-selectors https://www.example.com/news -o new_source.yaml

  # Append directly to sources.yaml (after review!)
  gocrawl generate-selectors https://www.example.com/news >> sources.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	GenerateCmd.Flags().IntVarP(&samples, "samples", "n", 1, "Number of sample articles to analyze (default: 1, future use)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourceURL := args[0]

	// Print to stderr (user feedback)
	fmt.Fprintf(os.Stderr, "🔍 Analyzing %s...\n\n", sourceURL)

	// Fetch the URL
	doc, err := fetchDocument(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Create discovery instance
	discovery, err := generator.NewSelectorDiscovery(doc, sourceURL)
	if err != nil {
		return fmt.Errorf("failed to create discovery instance: %w", err)
	}

	// Discover selectors
	result := discovery.DiscoverAll()

	// Print summary to stderr
	printSummary(os.Stderr, result)

	// Generate YAML
	yamlContent, err := generator.GenerateSourceYAML(sourceURL, result)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	// Output YAML
	var writer io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	_, err = fmt.Fprint(writer, yamlContent)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print success message to stderr
	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "\n✅ Selectors written to %s\n\n", outputFile)
		fmt.Fprintf(os.Stderr, "⚠️  IMPORTANT: Review and refine these selectors manually!\n")
		fmt.Fprintf(os.Stderr, "   After review, add to sources.yaml:\n")
		fmt.Fprintf(os.Stderr, "   cat %s >> sources.yaml\n", outputFile)
	} else {
		fmt.Fprintf(os.Stderr, "\n⚠️  IMPORTANT: Review and refine these selectors manually!\n")
	}

	return nil
}

// fetchDocument fetches a URL and returns a goquery document.
func fetchDocument(url string) (*goquery.Document, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

// printSummary prints a summary of discovered selectors to stderr.
func printSummary(w io.Writer, result generator.DiscoveryResult) {
	fmt.Fprintf(w, "📋 Discovered Selectors:\n\n")

	printCandidate(w, "title", result.Title)
	printCandidate(w, "body", result.Body)
	printCandidate(w, "author", result.Author)
	printCandidate(w, "published_time", result.PublishedTime)
	printCandidate(w, "image", result.Image)
	printCandidate(w, "link", result.Link)
	printCandidate(w, "category", result.Category)

	if len(result.Exclusions) > 0 {
		fmt.Fprintf(w, "\nexclude (%d patterns found):\n", len(result.Exclusions))
		for _, excl := range result.Exclusions {
			fmt.Fprintf(w, "  - %s\n", excl)
		}
	}

	fmt.Fprintf(w, "\n")
}

// printCandidate prints a selector candidate to stderr.
func printCandidate(w io.Writer, fieldName string, candidate generator.SelectorCandidate) {
	if len(candidate.Selectors) == 0 {
		return
	}

	fmt.Fprintf(w, "%s (confidence: %.0f%%):\n", fieldName, candidate.Confidence*100)
	for _, sel := range candidate.Selectors {
		fmt.Fprintf(w, "  - %s\n", sel)
	}
	if candidate.SampleText != "" {
		sample := candidate.SampleText
		if len(sample) > 80 {
			sample = sample[:80] + "..."
		}
		fmt.Fprintf(w, "  Sample: \"%s\"\n", sample)
	}
	fmt.Fprintf(w, "\n")
}

