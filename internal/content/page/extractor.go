// Package page provides functionality for processing and managing web pages.
package page

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
)

// extractText extracts text from the first element matching the selector.
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
			return strings.TrimSpace(text)
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
	// Apply exclude selectors before extraction
	applyExcludes(e, selectors.Exclude)

	data := &PageData{
		URL:       sourceURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate ID from URL
	data.ID = generateID(sourceURL)

	// Extract title using selector, with fallbacks
	data.Title = extractText(e, selectors.Title)
	if data.Title == "" {
		data.Title = extractMeta(e, "og:title")
	}
	if data.Title == "" {
		// Try to get from title tag
		data.Title = e.ChildText("title")
	}

	// Extract content using selector, with fallbacks
	data.Content = extractText(e, selectors.Content)
	if data.Content == "" {
		data.Content = extractText(e, "main")
	}
	if data.Content == "" {
		data.Content = extractText(e, "article")
	}
	if data.Content == "" {
		data.Content = extractText(e, "body")
	}

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

	// Extract Open Graph metadata using selectors, with fallbacks
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

	// Extract canonical URL using selector, with fallback
	data.CanonicalURL = extractAttr(e, selectors.Canonical, "href")
	if data.CanonicalURL == "" {
		data.CanonicalURL = extractAttr(e, "link[rel='canonical']", "href")
	}
	if data.CanonicalURL == "" {
		data.CanonicalURL = sourceURL
	}

	return data
}

// PageData holds extracted page data before conversion to models.Page
type PageData struct {
	ID           string
	URL          string
	Title        string
	Content      string
	Description  string
	Keywords     []string
	OgTitle      string
	OgDescription string
	OgImage      string
	OgURL        string
	CanonicalURL string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

