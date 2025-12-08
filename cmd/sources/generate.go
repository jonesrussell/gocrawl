// Package sources provides the sources command implementation.
package sources

import (
	"bufio"
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
	generateOutputFile   string
	generateArticleURL   string
	generateSamples      int
	generateShouldAppend bool
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
    -o new_source.yaml

  # Append directly to sources.yml (with confirmation and backup)
  gocrawl sources generate https://www.example.com/news \
    --article-url https://www.example.com/news/article-123 \
    --append`,
		Args: cobra.ExactArgs(1),
		RunE: runGenerate,
	}

	cmd.Flags().StringVarP(&generateOutputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&generateArticleURL, "article-url", "a", "", "Analyze an article page for better body/metadata selectors")
	cmd.Flags().IntVarP(&generateSamples, "samples", "n", 1, "Number of sample articles to analyze (default: 1, future use)")
	cmd.Flags().BoolVar(&generateShouldAppend, "append", false, "Append to sources.yml after confirmation (creates backup)")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourceURL := args[0]
	var backupFilePath string // Store backup file path for later reference

	// If output file is specified, ensure directory exists
	if generateOutputFile != "" && !generateShouldAppend {
		outputDir := filepath.Dir(generateOutputFile)
		if outputDir != "." && outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			fmt.Fprintf(os.Stderr, "📁 Created directory: %s\n", outputDir)
		}
	}

	// Handle --append flag
	if generateShouldAppend {
		generateOutputFile = "sources.yml"

		// Create backups directory
		if err := os.MkdirAll("backups", 0755); err != nil {
			return fmt.Errorf("failed to create backups directory: %w", err)
		}

		// Create timestamped backup
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupFilePath = filepath.Join("backups", fmt.Sprintf("sources_%s.yaml", timestamp))

		// Read and backup sources.yml if it exists
		if _, err := os.Stat("sources.yml"); err == nil {
			input, err := os.ReadFile("sources.yml")
			if err != nil {
				return fmt.Errorf("failed to read sources.yml: %w", err)
			}

			if err := os.WriteFile(backupFilePath, input, 0644); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}

			fmt.Fprintf(os.Stderr, "💾 Backup created: %s\n", backupFilePath)
		}

		// Prompt for confirmation
		fmt.Fprintf(os.Stderr, "\n⚠️  This will append to sources.yml. Continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(os.Stderr, "❌ Cancelled")
			return nil
		}
	}

	// Print to stderr (user feedback)
	fmt.Fprintf(os.Stderr, "🔍 Analyzing %s...\n", sourceURL)

	// Fetch the main page
	mainDoc, err := fetchDocument(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Create discovery instance for main page
	mainDiscovery, err := generator.NewSelectorDiscovery(mainDoc, sourceURL)
	if err != nil {
		return fmt.Errorf("failed to create discovery instance: %w", err)
	}

	// Discover selectors from main page
	mainResult := mainDiscovery.DiscoverAll()

	var finalResult generator.DiscoveryResult

	// If article URL provided, fetch and merge
	if generateArticleURL != "" {
		fmt.Fprintf(os.Stderr, "🔍 Analyzing article page %s...\n", generateArticleURL)
		articleDoc, err := fetchDocument(generateArticleURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to fetch article page: %v\n", err)
			fmt.Fprintf(os.Stderr, "   Continuing with main page results only...\n\n")
			finalResult = mainResult
		} else {
			articleDiscovery, err := generator.NewSelectorDiscovery(articleDoc, generateArticleURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Failed to create article discovery: %v\n", err)
				fmt.Fprintf(os.Stderr, "   Continuing with main page results only...\n\n")
				finalResult = mainResult
			} else {
				articleResult := articleDiscovery.DiscoverAll()

				// Merge results (article page takes precedence for content)
				finalResult = mergeResults(mainResult, articleResult)
				fmt.Fprintf(os.Stderr, "✅ Merged results from both pages\n\n")
			}
		}
	} else {
		finalResult = mainResult
	}

	// Print summary to stderr
	printSummary(os.Stderr, finalResult)

	// Check for missing fields and warn
	checkMissingFields(os.Stderr, finalResult)

	// Generate YAML
	yamlContent, err := generator.GenerateSourceYAML(sourceURL, finalResult)
	if err != nil {
		return fmt.Errorf("failed to generate YAML: %w", err)
	}

	// Check for duplicate source if appending
	if generateShouldAppend {
		// Parse the generated source to get its name
		var generatedSource struct {
			Name string `yaml:"name"`
		}
		// The YAML content starts with "  - name:", so we need to extract just the source entry
		// For simplicity, we'll search for the name pattern in the YAML
		lines := strings.Split(yamlContent, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "name:") {
				// Extract name value
				parts := strings.Split(line, "name:")
				if len(parts) > 1 {
					name := strings.Trim(strings.TrimSpace(parts[1]), "\"")
					generatedSource.Name = name
					break
				}
			}
		}

		// Check if source already exists in sources.yml
		if generatedSource.Name != "" {
			if _, err := os.Stat("sources.yml"); err == nil {
				existingContent, err := os.ReadFile("sources.yml")
				if err != nil {
					return fmt.Errorf("failed to read sources.yml: %w", err)
				}

				// Simple check - look for the source name
				searchPattern := fmt.Sprintf("name: \"%s\"", generatedSource.Name)
				if strings.Contains(string(existingContent), searchPattern) {
					fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: Source \"%s\" already exists in sources.yml!\n", generatedSource.Name)
					fmt.Fprintln(os.Stderr, "   Options:")
					fmt.Fprintln(os.Stderr, "   1. Cancel and manually merge (recommended)")
					fmt.Fprintln(os.Stderr, "   2. Append anyway (will create duplicate)")
					fmt.Fprintf(os.Stderr, "\n   Continue with append? [y/N]: ")

					reader := bufio.NewReader(os.Stdin)
					response, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read confirmation: %w", err)
					}

					response = strings.ToLower(strings.TrimSpace(response))
					if response != "y" && response != "yes" {
						fmt.Fprintln(os.Stderr, "\n❌ Cancelled. Please manually merge:")
						if generateOutputFile != "" && !generateShouldAppend {
							fmt.Fprintf(os.Stderr, "   code -d sources.yml %s\n", generateOutputFile)
						} else {
							fmt.Fprintln(os.Stderr, "   Review the generated YAML above and merge manually")
						}
						return nil
					}
				}
			}
		}
	}

	// Output YAML
	var writer io.Writer = os.Stdout
	if generateOutputFile != "" {
		var file *os.File
		var err error

		if generateShouldAppend {
			// Append mode
			file, err = os.OpenFile(generateOutputFile, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open sources.yml for appending: %w", err)
			}
		} else {
			// Overwrite mode
			file, err = os.Create(generateOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
		}
		defer file.Close()
		writer = file
	}

	_, err = fmt.Fprint(writer, yamlContent)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print success message to stderr
	if generateOutputFile != "" {
		if generateShouldAppend {
			fmt.Fprintf(os.Stderr, "\n✅ Appended to %s\n", generateOutputFile)
			if backupFilePath != "" {
				fmt.Fprintf(os.Stderr, "   To undo: cp %s %s\n", backupFilePath, generateOutputFile)
			}
		} else {
			fmt.Fprintf(os.Stderr, "\n✅ Selectors written to %s\n\n", generateOutputFile)
			fmt.Fprintf(os.Stderr, "⚠️  IMPORTANT: Review and refine these selectors manually!\n")
			fmt.Fprintf(os.Stderr, "   After review, add to sources.yml:\n")
			fmt.Fprintf(os.Stderr, "   cat %s >> sources.yml\n", generateOutputFile)
		}
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
