// Package sources provides the sources command implementation.
package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jonesrussell/gocrawl/internal/generator"
	"github.com/spf13/cobra"
)

var (
	generateOutputFile string
	generateArticleURL string
	generateSamples    int
)

// NewGenerateCommand creates a new generate subcommand for sources.
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [url]",
		Short: "Generate CSS selectors for a new source",
		Long: `Analyzes a news source and generates initial CSS selectors.

Example:
  # Write to file for review
  gocrawl sources generate https://www.example.com/news -o new_source.yaml

  # Analyze both listing and article pages for best results
  gocrawl sources generate https://www.example.com/news \
    --article-url https://www.example.com/news/article-123 \
    -o new_source.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: runGenerate,
	}

	cmd.Flags().StringVarP(&generateOutputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&generateArticleURL, "article-url", "a", "",
		"Analyze an article page for better body/metadata selectors")
	cmd.Flags().IntVarP(&generateSamples, "samples", "n", 1,
		"Number of sample articles to analyze (default: 1, future use)")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourceURL := args[0]

	// Prepare output directory if needed
	if err := prepareOutputDirectory(); err != nil {
		return err
	}

	// Discover selectors
	finalResult, err := discoverSelectors(sourceURL)
	if err != nil {
		return err
	}

	// Print summary and check for missing fields
	printSummary(os.Stderr, finalResult)
	checkMissingFields(os.Stderr, finalResult)

	// Generate YAML
	yamlContent, err := generator.GenerateSourceYAML(sourceURL, finalResult)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	// Write output
	if writeErr := writeOutput(yamlContent); writeErr != nil {
		return writeErr
	}

	// Print success message
	printSuccessMessage()

	return nil
}

// prepareOutputDirectory ensures the output directory exists if needed.
func prepareOutputDirectory() error {
	if generateOutputFile == "" {
		return nil
	}

	outputDir := filepath.Dir(generateOutputFile)
	if outputDir == "." || outputDir == "" {
		return nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	fmt.Fprintf(os.Stderr, "📁 Created directory: %s\n", outputDir)
	return nil
}

// discoverSelectors discovers selectors from the source URL and optionally an article URL.
func discoverSelectors(sourceURL string) (generator.DiscoveryResult, error) {
	fmt.Fprintf(os.Stderr, "🔍 Analyzing %s...\n", sourceURL)

	// Fetch the main page
	mainDoc, err := fetchDocument(sourceURL)
	if err != nil {
		return generator.DiscoveryResult{}, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Create discovery instance for main page
	mainDiscovery, err := generator.NewSelectorDiscovery(mainDoc, sourceURL)
	if err != nil {
		return generator.DiscoveryResult{}, fmt.Errorf("failed to create discovery instance: %w", err)
	}

	// Discover selectors from main page
	mainResult := mainDiscovery.DiscoverAll()

	// If article URL provided, fetch and merge
	if generateArticleURL == "" {
		return mainResult, nil
	}

	return discoverAndMergeArticleSelectors(generateArticleURL, mainResult)
}

// discoverAndMergeArticleSelectors fetches article page and merges results.
func discoverAndMergeArticleSelectors(
	articleURL string,
	mainResult generator.DiscoveryResult,
) (generator.DiscoveryResult, error) {
	fmt.Fprintf(os.Stderr, "🔍 Analyzing article page %s...\n", articleURL)

	articleDoc, fetchErr := fetchDocument(articleURL)
	if fetchErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to fetch article page: %v\n", fetchErr)
		fmt.Fprintf(os.Stderr, "   Continuing with main page results only...\n\n")
		return mainResult, nil
	}

	articleDiscovery, discoveryErr := generator.NewSelectorDiscovery(articleDoc, articleURL)
	if discoveryErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to create article discovery: %v\n", discoveryErr)
		fmt.Fprintf(os.Stderr, "   Continuing with main page results only...\n\n")
		return mainResult, nil
	}

	articleResult := articleDiscovery.DiscoverAll()
	finalResult := mergeResults(mainResult, articleResult)
	fmt.Fprintf(os.Stderr, "✅ Merged results from both pages\n\n")

	return finalResult, nil
}

// writeOutput writes YAML content to the appropriate output.
func writeOutput(yamlContent string) error {
	var writer io.Writer = os.Stdout

	if generateOutputFile == "" {
		_, err := fmt.Fprint(writer, yamlContent)
		return err
	}

	file, fileErr := os.Create(generateOutputFile)
	if fileErr != nil {
		return fmt.Errorf("failed to create output file: %w", fileErr)
	}
	defer file.Close()
	writer = file

	_, writeErr := fmt.Fprint(writer, yamlContent)
	if writeErr != nil {
		return fmt.Errorf("failed to write output: %w", writeErr)
	}

	return nil
}

// printSuccessMessage prints success message after writing output.
func printSuccessMessage() {
	if generateOutputFile == "" {
		fmt.Fprintf(os.Stderr, "\n⚠️  IMPORTANT: Review and refine these selectors manually!\n")
		return
	}

	fmt.Fprintf(os.Stderr, "\n✅ Selectors written to %s\n\n", generateOutputFile)
	fmt.Fprintf(os.Stderr, "⚠️  IMPORTANT: Review and refine these selectors manually!\n")
	fmt.Fprintf(os.Stderr, "   After review, use the sources API to add this source.\n")
}

// fetchDocument fetches a URL and returns a goquery document.
func fetchDocument(url string) (*goquery.Document, error) {
	ctx := context.Background()
	const httpTimeout = 30 * time.Second
	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a user agent
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	req.Header.Set("User-Agent", userAgent)

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

	const confidencePercent = 100.0
	fmt.Fprintf(w, "%s (confidence: %.0f%%):\n", fieldName, candidate.Confidence*confidencePercent)
	for _, sel := range candidate.Selectors {
		fmt.Fprintf(w, "  - %s\n", sel)
	}
	if candidate.SampleText != "" {
		sample := candidate.SampleText
		const maxSampleDisplayLength = 80
		if len(sample) > maxSampleDisplayLength {
			sample = sample[:maxSampleDisplayLength] + "..."
		}
		fmt.Fprintf(w, "  Sample: \"%s\"\n", sample)
	}
	fmt.Fprintf(w, "\n")
}

// checkMissingFields checks for missing critical fields and warns the user.
func checkMissingFields(w io.Writer, result generator.DiscoveryResult) {
	missingFields := []string{}
	fieldMap := map[string]generator.SelectorCandidate{
		"title":          result.Title,
		"body":           result.Body,
		"author":         result.Author,
		"published_time": result.PublishedTime,
		"image":          result.Image,
	}

	for field, candidate := range fieldMap {
		if len(candidate.Selectors) == 0 || candidate.Confidence == 0 {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		fmt.Fprintf(w, "⚠️  Missing fields: %s\n", strings.Join(missingFields, ", "))
		fmt.Fprintf(w, "   These will need to be added manually.\n\n")

		// Special warning for body field
		if contains(missingFields, "body") {
			fmt.Fprintf(w, "💡 TIP: No article body found!\n")
			fmt.Fprintf(w, "   This might be a listing page, not an article page.\n")
			fmt.Fprintf(w, "   Try running against an actual article URL for better results:\n")
			fmt.Fprintf(w, "   gocrawl sources generate <article-url> -o output.yaml\n\n")
		}
	}
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// mergeResults combines selectors from listing and article pages intelligently.
// Article page takes precedence for content fields, main page for structural fields.
func mergeResults(main, article generator.DiscoveryResult) generator.DiscoveryResult {
	merged := generator.DiscoveryResult{
		Exclusions: main.Exclusions, // Use main page exclusions (usually more comprehensive)
	}

	// Prefer article page for content fields (usually better on article pages)
	// Title: Use article if it has better confidence or main has none
	if article.Title.Confidence > main.Title.Confidence ||
		(len(main.Title.Selectors) == 0 && len(article.Title.Selectors) > 0) {
		merged.Title = article.Title
	} else {
		merged.Title = main.Title
	}

	// Body: Always prefer article page (listing pages rarely have article body)
	if len(article.Body.Selectors) > 0 {
		merged.Body = article.Body
	} else {
		merged.Body = main.Body
	}

	// Author: Prefer article page
	if len(article.Author.Selectors) > 0 {
		merged.Author = article.Author
	} else {
		merged.Author = main.Author
	}

	// PublishedTime: Prefer article page
	if len(article.PublishedTime.Selectors) > 0 {
		merged.PublishedTime = article.PublishedTime
	} else {
		merged.PublishedTime = main.PublishedTime
	}

	// Category: Use article if available, otherwise main
	if len(article.Category.Selectors) > 0 && article.Category.Confidence > main.Category.Confidence {
		merged.Category = article.Category
	} else {
		merged.Category = main.Category
	}

	// Image: Prefer main page (listing pages often have better featured images)
	if main.Image.Confidence > article.Image.Confidence ||
		(len(article.Image.Selectors) == 0 && len(main.Image.Selectors) > 0) {
		merged.Image = main.Image
	} else {
		merged.Image = article.Image
	}

	// Link: Always prefer main page (listing pages have article links)
	if len(main.Link.Selectors) > 0 {
		merged.Link = main.Link
	} else {
		merged.Link = article.Link
	}

	return merged
}
