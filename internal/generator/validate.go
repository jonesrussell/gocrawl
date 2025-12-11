// Package generator provides tools for generating CSS selector configurations
// for news sources.
package generator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
)

// ValidationResult holds the results of validating selectors against articles.
type ValidationResult struct {
	// FieldResults maps field names to their validation results
	FieldResults map[string]FieldValidationResult
	// TotalArticles is the number of articles tested
	TotalArticles int
	// SuccessfulArticles is the number of articles where all critical fields were found
	SuccessfulArticles int
}

// FieldValidationResult holds validation results for a single field.
type FieldValidationResult struct {
	// FieldName is the name of the field
	FieldName string
	// SuccessCount is the number of articles where this field was found
	SuccessCount int
	// TotalCount is the total number of articles tested
	TotalCount int
	// SuccessRate is the percentage of articles where this field was found (0-100)
	SuccessRate float64
	// FailedURLs is a list of URLs where this field was not found
	FailedURLs []string
	// SampleValues contains sample extracted values for verification
	SampleValues []string
}

// ValidateSelectors validates selectors against a list of article URLs.
func ValidateSelectors(
	selectors configtypes.ArticleSelectors,
	articleURLs []string,
	maxSamples int,
) (*ValidationResult, error) {
	if len(articleURLs) == 0 {
		return nil, errors.New("no article URLs provided")
	}

	if len(articleURLs) > maxSamples {
		articleURLs = articleURLs[:maxSamples]
	}

	result := &ValidationResult{
		FieldResults:  make(map[string]FieldValidationResult),
		TotalArticles: len(articleURLs),
	}

	fields := buildFieldMap(selectors)
	initializeFieldResults(result, fields, len(articleURLs))

	criticalFields := []string{"title", "body"}
	articlesWithAllCritical := validateArticles(articleURLs, fields, result, criticalFields)

	result.SuccessfulArticles = articlesWithAllCritical
	return result, nil
}

// buildFieldMap creates a map of field names to selectors.
func buildFieldMap(selectors configtypes.ArticleSelectors) map[string]string {
	return map[string]string{
		"title":          selectors.Title,
		"body":           selectors.Body,
		"author":         selectors.Author,
		"byline":         selectors.Byline,
		"published_time": selectors.PublishedTime,
		"image":          selectors.Image,
		"link":           selectors.Link,
		"category":       selectors.Category,
		"section":        selectors.Section,
	}
}

// initializeFieldResults initializes field results for all fields.
func initializeFieldResults(result *ValidationResult, fields map[string]string, totalCount int) {
	for fieldName := range fields {
		result.FieldResults[fieldName] = FieldValidationResult{
			FieldName:    fieldName,
			SuccessCount: 0,
			TotalCount:   totalCount,
			FailedURLs:   []string{},
			SampleValues: []string{},
		}
	}
}

// validateArticles validates selectors against all article URLs.
func validateArticles(
	articleURLs []string,
	fields map[string]string,
	result *ValidationResult,
	criticalFields []string,
) int {
	articlesWithAllCritical := 0

	for _, articleURL := range articleURLs {
		doc, err := fetchDocumentForValidation(articleURL)
		if err != nil {
			markAllFieldsFailed(result, fields, articleURL)
			continue
		}

		if validateArticleFields(doc, fields, result, articleURL, criticalFields) {
			articlesWithAllCritical++
		}
	}

	return articlesWithAllCritical
}

// markAllFieldsFailed marks all fields as failed for a URL.
func markAllFieldsFailed(result *ValidationResult, fields map[string]string, articleURL string) {
	for fieldName := range fields {
		fieldResult := result.FieldResults[fieldName]
		fieldResult.FailedURLs = append(fieldResult.FailedURLs, articleURL)
		result.FieldResults[fieldName] = fieldResult
	}
}

// validateArticleFields validates all fields for a single article.
// Returns true if article has all critical fields.
func validateArticleFields(
	doc *goquery.Document,
	fields map[string]string,
	result *ValidationResult,
	articleURL string,
	criticalFields []string,
) bool {
	articleHasAllCritical := true

	for fieldName, selector := range fields {
		if selector == "" {
			continue
		}

		value, found := extractValueFromDocument(doc, selector)
		fieldResult := result.FieldResults[fieldName]

		if found && value != "" {
			updateFieldSuccess(&fieldResult, value)
		} else {
			updateFieldFailure(&fieldResult, articleURL, fieldName, criticalFields, &articleHasAllCritical)
		}

		updateFieldSuccessRate(&fieldResult, result.TotalArticles)
		result.FieldResults[fieldName] = fieldResult
	}

	return articleHasAllCritical
}

// updateFieldSuccess updates field result for successful extraction.
func updateFieldSuccess(fieldResult *FieldValidationResult, value string) {
	fieldResult.SuccessCount++

	const maxSamples = 3
	const maxSampleLength = 100
	if len(fieldResult.SampleValues) >= maxSamples {
		return
	}

	sample := value
	if len(sample) > maxSampleLength {
		sample = sample[:maxSampleLength] + "..."
	}
	fieldResult.SampleValues = append(fieldResult.SampleValues, sample)
}

// updateFieldFailure updates field result for failed extraction.
func updateFieldFailure(
	fieldResult *FieldValidationResult,
	articleURL string,
	fieldName string,
	criticalFields []string,
	articleHasAllCritical *bool,
) {
	fieldResult.FailedURLs = append(fieldResult.FailedURLs, articleURL)
	if contains(criticalFields, fieldName) {
		*articleHasAllCritical = false
	}
}

// updateFieldSuccessRate calculates and updates the success rate.
func updateFieldSuccessRate(fieldResult *FieldValidationResult, totalCount int) {
	const percentMultiplier = 100.0
	fieldResult.SuccessRate = float64(fieldResult.SuccessCount) / float64(totalCount) * percentMultiplier
}

// extractValueFromDocument extracts a value from a goquery document using a selector.
func extractValueFromDocument(doc *goquery.Document, selector string) (string, bool) {
	if selector == "" {
		return "", false
	}

	// Handle meta tags
	if strings.HasPrefix(selector, "meta[") {
		return extractMetaContent(doc, selector)
	}

	// Handle attributes (e.g., time[datetime], img[src])
	if strings.Contains(selector, "[") {
		return extractAttributeValue(doc, selector)
	}

	// Regular text extraction
	return extractTextValue(doc, selector)
}

// extractMetaContent extracts content from meta tags.
func extractMetaContent(doc *goquery.Document, selector string) (string, bool) {
	selection := doc.Find(selector).First()
	if selection.Length() == 0 {
		return "", false
	}
	content, exists := selection.Attr("content")
	if exists && content != "" {
		return strings.TrimSpace(content), true
	}
	return "", false
}

// extractAttributeValue extracts value from elements with attributes.
func extractAttributeValue(doc *goquery.Document, selector string) (string, bool) {
	parts := strings.Split(selector, "[")
	if len(parts) <= 1 {
		return "", false
	}

	attrPart := strings.TrimSuffix(parts[1], "]")
	attrParts := strings.Split(attrPart, "=")
	if len(attrParts) == 0 {
		return "", false
	}

	attrName := strings.Trim(attrParts[0], "'\"")
	if attrName != "datetime" && attrName != "src" && attrName != "href" {
		return "", false
	}

	elementSelector := parts[0]
	selection := doc.Find(elementSelector).First()
	if selection.Length() == 0 {
		return "", false
	}

	value, exists := selection.Attr(attrName)
	if exists && value != "" {
		return strings.TrimSpace(value), true
	}

	return "", false
}

// extractTextValue extracts text content from selectors.
func extractTextValue(doc *goquery.Document, selector string) (string, bool) {
	selectors := strings.Split(selector, ",")
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}
		selection := doc.Find(sel).First()
		if selection.Length() > 0 {
			text := strings.TrimSpace(selection.Text())
			if text != "" {
				return text, true
			}
		}
	}
	return "", false
}

// FetchDocumentForValidation fetches a document for validation.
// Exported for use by validate command.
func FetchDocumentForValidation(url string) (*goquery.Document, error) {
	return fetchDocumentForValidation(url)
}

// fetchDocumentForValidation fetches a document for validation.
func fetchDocumentForValidation(url string) (*goquery.Document, error) {
	const requestTimeout = 30 * time.Second
	client := &http.Client{
		Timeout: requestTimeout,
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
