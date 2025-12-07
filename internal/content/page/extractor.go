// Package page provides functionality for processing and managing web pages.
package page

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
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

// extractPage extracts page data from HTML element.
// For pages, we use simpler extraction since pages don't have article-specific selectors.
func extractPage(e *colly.HTMLElement, sourceURL string) *PageData {
	data := &PageData{
		URL:       sourceURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate ID from URL
	data.ID = generateID(sourceURL)

	// Extract title - try h1, then og:title, then title tag
	data.Title = extractText(e, "h1")
	if data.Title == "" {
		data.Title = extractMeta(e, "og:title")
	}
	if data.Title == "" {
		// Try to get from title tag
		data.Title = e.ChildText("title")
	}

	// Extract content - try main, article, or body
	data.Content = extractText(e, "main")
	if data.Content == "" {
		data.Content = extractText(e, "article")
	}
	if data.Content == "" {
		data.Content = extractText(e, "body")
	}

	// Extract description
	data.Description = extractMetaName(e, "description")
	if data.Description == "" {
		data.Description = extractMeta(e, "og:description")
	}

	// Extract keywords
	keywordsStr := extractMetaName(e, "keywords")
	if keywordsStr != "" {
		data.Keywords = strings.Split(keywordsStr, ",")
		for i := range data.Keywords {
			data.Keywords[i] = strings.TrimSpace(data.Keywords[i])
		}
	}

	// Extract Open Graph metadata
	data.OgTitle = extractMeta(e, "og:title")
	if data.OgTitle == "" {
		data.OgTitle = data.Title
	}

	data.OgDescription = extractMeta(e, "og:description")
	if data.OgDescription == "" {
		data.OgDescription = data.Description
	}

	data.OgImage = extractMeta(e, "og:image")
	data.OgURL = extractMeta(e, "og:url")

	// Extract canonical URL
	data.CanonicalURL = extractAttr(e, "link[rel='canonical']", "href")
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

