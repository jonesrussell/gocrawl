// Package articles provides functionality for processing and managing article content.
package articles

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
// Returns empty string if not found (Colly returns empty string safely).
// Uses DOM.Find() to search anywhere in the element, not just direct children.
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
		// First try ChildText (for direct children, faster)
		text := e.ChildText(sel)
		if text != "" {
			return strings.TrimSpace(text)
		}
		// If ChildText didn't find it, try DOM.Find() to search anywhere
		element := e.DOM.Find(sel).First()
		if element.Length() > 0 {
			text = element.Text()
			if text != "" {
				return strings.TrimSpace(text)
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
			cleaned := strings.TrimSpace(text)
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

// extractAttr extracts an attribute from the first element matching the selector.
// Returns empty string if not found.
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

// parseDate attempts to parse a date string in various formats.
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	// Common date formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC850,
		time.RFC850,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05+07:00",
	}

	dateStr = strings.TrimSpace(dateStr)
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Try parsing as Unix timestamp
	if unixTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return unixTime
	}

	return time.Time{}
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

// extractArticle extracts article data from HTML element using selectors.
func extractArticle(e *colly.HTMLElement, selectors configtypes.ArticleSelectors, sourceURL string) *ArticleData {
	data := &ArticleData{
		Source:    sourceURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Extract basic fields (before applying excludes, as these are usually in head or specific locations)
	data.Title = extractText(e, selectors.Title)
	if data.Title == "" {
		// Fallback to OG title
		data.Title = extractMeta(e, "og:title")
	}

	// Extract body - use container-based extraction if container selector is available
	// This is the most reliable method as it scopes to the article container and applies excludes
	if selectors.Container != "" {
		// Use container-based extraction with excludes applied
		data.Body = extractTextFromContainer(e, selectors.Container, selectors.Exclude)
		// If container extraction didn't work, fall back to body selector with excludes
		if data.Body == "" {
			// Apply excludes to the element before extracting body
			applyExcludes(e, selectors.Exclude)
			data.Body = extractText(e, selectors.Body)
		}
	} else {
		// No container selector, apply excludes and use body selector directly
		applyExcludes(e, selectors.Exclude)
		data.Body = extractText(e, selectors.Body)
	}

	// Additional fallbacks for body if still empty
	if data.Body == "" {
		// Try common article content containers
		data.Body = extractTextFromContainer(e, "article, main, .article-content, .article-body", selectors.Exclude)
	}

	data.Intro = extractText(e, selectors.Intro)
	if data.Intro == "" {
		// Fallback to OG description
		data.Intro = extractMeta(e, "og:description")
	}

	// Extract metadata
	data.Author = extractText(e, selectors.Author)
	if data.Author == "" {
		data.Author = extractMeta(e, "article:author")
	}

	data.BylineName = extractText(e, selectors.BylineName)
	if data.BylineName == "" {
		data.BylineName = extractText(e, selectors.Byline)
	}

	// Extract dates
	publishedTimeStr := extractAttr(e, selectors.PublishedTime, "datetime")
	if publishedTimeStr == "" {
		publishedTimeStr = extractText(e, selectors.PublishedTime)
	}
	if publishedTimeStr == "" {
		publishedTimeStr = extractMeta(e, "article:published_time")
	}
	if publishedTimeStr != "" {
		data.PublishedDate = parseDate(publishedTimeStr)
	}

	// Extract tags/keywords
	keywordsStr := extractText(e, selectors.Keywords)
	if keywordsStr == "" {
		keywordsStr = extractMetaName(e, "keywords")
	}
	if keywordsStr != "" {
		data.Tags = strings.Split(keywordsStr, ",")
		for i := range data.Tags {
			data.Tags[i] = strings.TrimSpace(data.Tags[i])
		}
	}

	// Extract tags from tags selector
	if tagsStr := extractText(e, selectors.Tags); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				// Avoid duplicates
				found := false
				for _, existing := range data.Tags {
					if existing == tag {
						found = true
						break
					}
				}
				if !found {
					data.Tags = append(data.Tags, tag)
				}
			}
		}
	}

	// Extract Open Graph metadata
	data.OgTitle = extractMeta(e, "og:title")
	if data.OgTitle == "" {
		data.OgTitle = data.Title
	}

	data.OgDescription = extractMeta(e, "og:description")
	if data.OgDescription == "" {
		data.OgDescription = data.Intro
	}

	data.OgImage = extractMeta(e, "og:image")
	data.OgURL = extractMeta(e, "og:url")
	data.OgType = extractMeta(e, "og:type")
	data.OgSiteName = extractMeta(e, "og:site_name")

	// Extract other metadata
	data.Description = extractMetaName(e, "description")
	if data.Description == "" {
		data.Description = data.Intro
	}

	data.Section = extractText(e, selectors.Section)
	if data.Section == "" {
		data.Section = extractMeta(e, "article:section")
	}

	data.Category = extractText(e, selectors.Category)
	if data.Category == "" {
		data.Category = extractMeta(e, "article:section")
	}

	data.CanonicalURL = extractAttr(e, selectors.Canonical, "href")
	if data.CanonicalURL == "" {
		data.CanonicalURL = sourceURL
	}

	// Extract article ID if available
	articleID := extractAttr(e, selectors.ArticleID, "data-article-id")
	if articleID == "" {
		articleID = extractAttr(e, selectors.ArticleID, "data-post-id")
	}
	if articleID == "" {
		articleID = extractAttr(e, selectors.ArticleID, "id")
	}

	// Generate ID from URL if article ID not found
	if articleID == "" {
		articleID = generateID(sourceURL)
	}
	data.ID = articleID

	return data
}

// ArticleData holds extracted article data before conversion to models.Article
type ArticleData struct {
	ID            string
	Title         string
	Body          string
	Intro         string
	Author        string
	BylineName    string
	PublishedDate time.Time
	Source        string
	Tags          []string
	Description   string
	Section       string
	Category      string
	OgTitle       string
	OgDescription string
	OgImage       string
	OgURL         string
	OgType        string
	OgSiteName    string
	CanonicalURL  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

