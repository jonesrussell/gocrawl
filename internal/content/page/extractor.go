// Package page provides functionality for processing and managing web pages.
package page

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
)

var (
	// JavaScript patterns to remove from extracted text
	jsPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)document\.addEventListener[^)]*\)`),
		regexp.MustCompile(`(?i)function\s*\([^)]*\)\s*\{[^}]*\}`),
		regexp.MustCompile(`(?i)\.replaceWith\([^)]*\)`),
		regexp.MustCompile(`(?i)\.cloneNode\([^)]*\)`),
		regexp.MustCompile(`(?i)template\.content`),
		regexp.MustCompile(`(?i)\.dataset\.[a-zA-Z]+`),
		regexp.MustCompile(`(?i)\.parentElement`),
		regexp.MustCompile(`(?i)getElementById\([^)]*\)`),
		regexp.MustCompile(`(?i)querySelector\([^)]*\)`),
	}
	// Whitespace normalization regex
	whitespaceRegex = regexp.MustCompile(`\s+`)
	// Multiple newlines regex
	newlineRegex = regexp.MustCompile(`\n{3,}`)
)

// cleanText removes JavaScript code, normalizes whitespace, and cleans extracted text.
func cleanText(text string) string {
	if text == "" {
		return ""
	}

	// Remove JavaScript patterns
	for _, pattern := range jsPatterns {
		text = pattern.ReplaceAllString(text, "")
	}

	// Normalize whitespace (replace multiple spaces/tabs with single space)
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Replace multiple newlines with double newline
	text = newlineRegex.ReplaceAllString(text, "\n\n")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}

// extractText extracts text from the first element matching the selector.
// If a container selector is provided, it extracts from within that container.
func extractText(e *colly.HTMLElement, selector string) string {
	if selector == "" {
		return ""
	}
	// Try each selector if comma-separated
	selectors := strings.Split(selector, ",")
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}
		text := e.ChildText(sel)
		if text != "" {
			cleaned := cleanText(text)
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

// extractTextFromContainer extracts text from a container element, applying excludes first.
func extractTextFromContainer(e *colly.HTMLElement, containerSelector string, excludes []string) string {
	if containerSelector == "" {
		return ""
	}

	// Try each container selector if comma-separated
	selectors := strings.Split(containerSelector, ",")
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}

		// Find the container element
		container := e.DOM.Find(sel).First()
		if container.Length() == 0 {
			continue
		}

		// Apply exclude patterns to the container
		for _, excludeSelector := range excludes {
			if excludeSelector != "" {
				container.Find(excludeSelector).Remove()
			}
		}

		// Extract text from the cleaned container
		text := container.Text()
		if text != "" {
			cleaned := cleanText(text)
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

// extractAttr extracts an attribute from the first element matching the selector.
func extractAttr(e *colly.HTMLElement, selector, attr string) string {
	if selector == "" || attr == "" {
		return ""
	}
	// Try each selector if comma-separated
	selectors := strings.Split(selector, ",")
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}
		value := e.ChildAttr(sel, attr)
		if value != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// extractMeta extracts a meta tag value using property attribute.
func extractMeta(e *colly.HTMLElement, property string) string {
	if property == "" {
		return ""
	}
	selector := fmt.Sprintf("meta[property='%s']", property)
	return e.ChildAttr(selector, "content")
}

// extractMetaName extracts a meta tag value using name attribute.
func extractMetaName(e *colly.HTMLElement, name string) string {
	if name == "" {
		return ""
	}
	selector := fmt.Sprintf("meta[name='%s']", name)
	return e.ChildAttr(selector, "content")
}

// generateID generates a unique ID from a URL using SHA256 hash.
func generateID(url string) string {
	if url == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// applyExcludes removes elements matching exclude selectors from the HTML element.
func applyExcludes(e *colly.HTMLElement, excludes []string) {
	for _, excludeSelector := range excludes {
		if excludeSelector != "" {
			e.DOM.Find(excludeSelector).Remove()
		}
	}
}

// GetSelectorsForURL retrieves the appropriate PageSelectors for a given URL.
func GetSelectorsForURL(sourceManager interface {
	FindSourceByURL(rawURL string) *configtypes.Source
}, url string) configtypes.PageSelectors {
	if sourceManager == nil {
		var defaultSelectors configtypes.PageSelectors
		return defaultSelectors.Default()
	}

	sourceConfig := sourceManager.FindSourceByURL(url)
	if sourceConfig != nil {
		return sourceConfig.Selectors.Page
	}

	var defaultSelectors configtypes.PageSelectors
	return defaultSelectors.Default()
}

// extractPage extracts page data from HTML element using selectors.
func extractPage(e *colly.HTMLElement, selectors configtypes.PageSelectors, sourceURL string) *PageData {
	data := &PageData{
		URL:       sourceURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate ID from URL
	data.ID = generateID(sourceURL)

	// Extract title
	extractPageTitle(data, e, selectors)

	// Extract content
	extractPageContent(data, e, selectors)

	// Extract description and keywords
	extractPageDescriptionKeywords(data, e, selectors)

	// Extract Open Graph metadata
	extractPageOpenGraphMetadata(data, e, selectors)

	// Extract canonical URL
	extractPageCanonicalURL(data, e, selectors, sourceURL)

	return data
}

// extractPageTitle extracts the page title with fallbacks.
func extractPageTitle(data *PageData, e *colly.HTMLElement, selectors configtypes.PageSelectors) {
	data.Title = extractText(e, selectors.Title)
	if data.Title == "" {
		data.Title = extractMeta(e, "og:title")
	}
	if data.Title == "" {
		// Try to get from title tag
		titleText := e.ChildText("title")
		data.Title = cleanText(titleText)
	}
}

// extractPageContent extracts the page content with multiple fallback strategies.
func extractPageContent(data *PageData, e *colly.HTMLElement, selectors configtypes.PageSelectors) {
	// Extract content - use container selector if available, otherwise use content selector
	if selectors.Container != "" {
		// Use container-based extraction with excludes applied
		data.Content = extractTextFromContainer(e, selectors.Container, selectors.Exclude)
	}

	// Fallback to content selector if container extraction didn't work
	if data.Content == "" {
		// Apply excludes to the entire element before extracting
		applyExcludes(e, selectors.Exclude)
		data.Content = extractText(e, selectors.Content)
	}

	// Additional fallbacks if still empty
	if data.Content == "" {
		// Try common content containers
		data.Content = extractTextFromContainer(e, "main", selectors.Exclude)
	}
	if data.Content == "" {
		data.Content = extractTextFromContainer(e, "article", selectors.Exclude)
	}
	if data.Content == "" {
		// Last resort: extract from body but apply excludes
		applyExcludes(e, selectors.Exclude)
		bodyText := e.ChildText("body")
		data.Content = cleanText(bodyText)
	}
}

// extractPageDescriptionKeywords extracts description and keywords.
func extractPageDescriptionKeywords(data *PageData, e *colly.HTMLElement, selectors configtypes.PageSelectors) {
	// Extract description using selector, with fallbacks
	data.Description = extractText(e, selectors.Description)
	if data.Description == "" {
		data.Description = extractMetaName(e, "description")
	}
	if data.Description == "" {
		data.Description = extractMeta(e, "og:description")
	}

	// Extract keywords using selector, with fallbacks
	keywordsStr := extractText(e, selectors.Keywords)
	if keywordsStr == "" {
		keywordsStr = extractMetaName(e, "keywords")
	}
	if keywordsStr != "" {
		data.Keywords = strings.Split(keywordsStr, ",")
		for i := range data.Keywords {
			data.Keywords[i] = strings.TrimSpace(data.Keywords[i])
		}
	}
}

// extractPageOpenGraphMetadata extracts Open Graph metadata.
func extractPageOpenGraphMetadata(data *PageData, e *colly.HTMLElement, selectors configtypes.PageSelectors) {
	data.OgTitle = extractText(e, selectors.OGTitle)
	if data.OgTitle == "" {
		data.OgTitle = extractMeta(e, "og:title")
	}
	if data.OgTitle == "" {
		data.OgTitle = data.Title
	}

	data.OgDescription = extractText(e, selectors.OGDescription)
	if data.OgDescription == "" {
		data.OgDescription = extractMeta(e, "og:description")
	}
	if data.OgDescription == "" {
		data.OgDescription = data.Description
	}

	data.OgImage = extractText(e, selectors.OGImage)
	if data.OgImage == "" {
		data.OgImage = extractMeta(e, "og:image")
	}

	data.OgURL = extractText(e, selectors.OgURL)
	if data.OgURL == "" {
		data.OgURL = extractMeta(e, "og:url")
	}
}

// extractPageCanonicalURL extracts the canonical URL.
func extractPageCanonicalURL(
	data *PageData,
	e *colly.HTMLElement,
	selectors configtypes.PageSelectors,
	sourceURL string,
) {
	data.CanonicalURL = extractAttr(e, selectors.Canonical, "href")
	if data.CanonicalURL == "" {
		data.CanonicalURL = extractAttr(e, "link[rel='canonical']", "href")
	}
	if data.CanonicalURL == "" {
		data.CanonicalURL = sourceURL
	}
}

// PageData holds extracted page data before conversion to models.Page
type PageData struct {
	ID            string
	URL           string
	Title         string
	Content       string
	Description   string
	Keywords      []string
	OgTitle       string
	OgDescription string
	OgImage       string
	OgURL         string
	CanonicalURL  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
