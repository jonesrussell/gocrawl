// Package sources provides the sources command implementation.
package sources

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jonesrussell/gocrawl/cmd/common"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/generator"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
)

var (
	validateSourceName string
	validateSamples    int
	validateURLs       []string
)

// NewValidateCommand creates a new validate subcommand for sources.
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate CSS selectors against real articles",
		Long: `Tests CSS selectors from a source configuration against real article URLs
to verify they work correctly.

Example:
  # Validate selectors for a source (fetches sample articles from source URL)
  gocrawl sources validate --source "Mid-North Monitor" --samples 5

  # Validate selectors against specific URLs
  gocrawl sources validate --source "Mid-North Monitor" ` +
			`--urls "https://example.com/article1" "https://example.com/article2"`,
		RunE: runValidate,
	}

	cmd.Flags().StringVarP(&validateSourceName, "source", "s", "", "Source name to validate (required)")
	const defaultSamples = 5
	cmd.Flags().IntVarP(&validateSamples, "samples", "n", defaultSamples,
		"Number of sample articles to test (default: 5)")
	cmd.Flags().StringSliceVarP(&validateURLs, "urls", "u", []string{},
		"Specific article URLs to test (overrides samples)")
	if err := cmd.MarkFlagRequired("source"); err != nil {
		return nil
	}

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Get dependencies
	deps, err := common.NewCommandDeps()
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Load sources
	sourceManager, err := sources.LoadSources(deps.Config, deps.Logger)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Find source by name
	sourceConfig := sourceManager.FindByName(validateSourceName)
	if sourceConfig == nil {
		return fmt.Errorf("source not found: %s", validateSourceName)
	}

	// Convert to configtypes.Source for selectors
	allSources, err := sourceManager.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	var articleSelectors configtypes.ArticleSelectors
	for i := range allSources {
		src := &allSources[i]
		if src.Name == validateSourceName {
			// Convert selectors
			articleSelectors = configtypes.ArticleSelectors{
				Container:     src.Selectors.Article.Container,
				Title:         src.Selectors.Article.Title,
				Body:          src.Selectors.Article.Body,
				Intro:         src.Selectors.Article.Intro,
				Link:          src.Selectors.Article.Link,
				Image:         src.Selectors.Article.Image,
				Author:        src.Selectors.Article.Author,
				Byline:        src.Selectors.Article.Byline,
				PublishedTime: src.Selectors.Article.PublishedTime,
				TimeAgo:       src.Selectors.Article.TimeAgo,
				Section:       src.Selectors.Article.Section,
				Category:      src.Selectors.Article.Category,
				ArticleID:     src.Selectors.Article.ArticleID,
				Exclude:       src.Selectors.Article.Exclude,
			}
			break
		}
	}

	// Get article URLs
	articleURLs := validateURLs
	if len(articleURLs) == 0 {
		// Try to discover article URLs from the source URL
		discoveredURLs, discoverErr := discoverArticleURLs(sourceConfig.URL, articleSelectors, validateSamples)
		if discoverErr != nil {
			return fmt.Errorf("failed to discover article URLs: %w\n   Please provide URLs with --urls flag", discoverErr)
		}
		if len(discoveredURLs) == 0 {
			return errors.New("no article URLs found on source page\n   Please provide URLs with --urls flag")
		}
		articleURLs = discoveredURLs
		fmt.Fprintf(os.Stderr, "📋 Discovered %d article URL(s) from source page\n", len(articleURLs))
	}

	// Validate selectors
	fmt.Fprintf(os.Stderr, "🧪 Testing selectors for \"%s\"...\n", validateSourceName)
	fmt.Fprintf(os.Stderr, "📄 Testing %d article(s)...\n\n", len(articleURLs))

	result, err := generator.ValidateSelectors(articleSelectors, articleURLs, validateSamples)
	if err != nil {
		return fmt.Errorf("failed to validate selectors: %w", err)
	}

	// Print results
	printValidationResults(os.Stderr, result)

	return nil
}

// printValidationResults prints validation results in a user-friendly format.
func printValidationResults(w *os.File, result *generator.ValidationResult) {
	printValidationHeader(w, result)
	printFieldResults(w, result)
	printValidationSummary(w, result)
}

// printValidationHeader prints the header with summary statistics.
func printValidationHeader(w *os.File, result *generator.ValidationResult) {
	fmt.Fprintf(w, "📊 Validation Results:\n\n")
	const percentMultiplier = 100.0
	fmt.Fprintf(w, "Total articles tested: %d\n", result.TotalArticles)
	fmt.Fprintf(w, "Articles with all critical fields: %d (%.0f%%)\n\n",
		result.SuccessfulArticles,
		float64(result.SuccessfulArticles)/float64(result.TotalArticles)*percentMultiplier)
}

// printFieldResults prints results for each field.
func printFieldResults(w *os.File, result *generator.ValidationResult) {
	fieldOrder := []string{"title", "body", "author", "byline", "published_time", "image", "link", "category", "section"}

	for _, fieldName := range fieldOrder {
		fieldResult, exists := result.FieldResults[fieldName]
		if !exists || fieldResult.TotalCount == 0 {
			continue
		}

		printFieldResult(w, fieldName, fieldResult)
	}
}

// printFieldResult prints a single field's validation result.
func printFieldResult(w *os.File, fieldName string, fieldResult generator.FieldValidationResult) {
	status := getStatusEmoji(fieldResult.SuccessRate)
	fmt.Fprintf(w, "%s %s: %.0f%% (%d/%d)\n",
		status,
		fieldName,
		fieldResult.SuccessRate,
		fieldResult.SuccessCount,
		fieldResult.TotalCount,
	)

	printSampleValues(w, fieldResult.SampleValues)
	printFailedURLs(w, fieldResult.FailedURLs)
	fmt.Fprintf(w, "\n")
}

// getStatusEmoji returns the appropriate emoji based on success rate.
func getStatusEmoji(successRate float64) string {
	const highSuccessRate = 90.0
	const mediumSuccessRate = 70.0

	if successRate >= highSuccessRate {
		return "✅"
	}
	if successRate >= mediumSuccessRate {
		return "⚠️"
	}
	return "❌"
}

// printSampleValues prints sample extracted values.
func printSampleValues(w *os.File, sampleValues []string) {
	const maxSamplesToShow = 2
	const maxSampleDisplayLength = 60

	if len(sampleValues) == 0 {
		return
	}

	for i, sample := range sampleValues {
		if i >= maxSamplesToShow {
			break
		}
		sampleDisplay := sample
		if len(sampleDisplay) > maxSampleDisplayLength {
			sampleDisplay = sampleDisplay[:maxSampleDisplayLength] + "..."
		}
		fmt.Fprintf(w, "   Sample %d: \"%s\"\n", i+1, sampleDisplay)
	}
}

// printFailedURLs prints failed URLs if any.
func printFailedURLs(w *os.File, failedURLs []string) {
	const maxFailedURLsToShow = 3

	if len(failedURLs) == 0 {
		return
	}

	if len(failedURLs) <= maxFailedURLsToShow {
		fmt.Fprintf(w, "   Failed on: %s\n", strings.Join(failedURLs, ", "))
		return
	}

	fmt.Fprintf(w, "   Failed on %d URLs (showing first %d): %s\n",
		len(failedURLs),
		maxFailedURLsToShow,
		strings.Join(failedURLs[:maxFailedURLsToShow], ", "))
}

// printValidationSummary prints the final summary.
func printValidationSummary(w *os.File, result *generator.ValidationResult) {
	fmt.Fprintf(w, "---\n\n")
	if result.SuccessfulArticles == result.TotalArticles {
		fmt.Fprintf(w, "✅ All articles have all critical fields!\n")
	} else {
		fmt.Fprintf(w, "⚠️  Some articles are missing critical fields.\n")
		fmt.Fprintf(w, "   Review failed URLs above and refine selectors if needed.\n")
	}
}

// discoverArticleURLs discovers article URLs from a source page using link selectors.
func discoverArticleURLs(sourceURL string, selectors configtypes.ArticleSelectors, maxSamples int) ([]string, error) {
	doc, err := generator.FetchDocumentForValidation(sourceURL)
	if err != nil {
		return nil, err
	}

	if selectors.Link == "" {
		return nil, nil
	}

	baseURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	return extractURLsFromSelectors(doc, selectors.Link, baseURL, maxSamples), nil
}

// extractURLsFromSelectors extracts URLs from comma-separated selectors.
func extractURLsFromSelectors(doc *goquery.Document, linkSelector string, baseURL *url.URL, maxSamples int) []string {
	var articleURLs []string
	linkSelectors := strings.Split(linkSelector, ",")

	for _, selector := range linkSelectors {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			continue
		}

		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			if len(articleURLs) >= maxSamples {
				return
			}

			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			hrefURL, err := baseURL.Parse(href)
			if err == nil {
				articleURLs = append(articleURLs, hrefURL.String())
			}
		})

		if len(articleURLs) >= maxSamples {
			break
		}
	}

	return articleURLs
}
