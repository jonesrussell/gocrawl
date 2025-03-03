// Package content handles non-article content operations
package content

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// Interface defines the interface for content operations
type Interface interface {
	ExtractContent(e *colly.HTMLElement) *models.Content
	ExtractMetadata(e *colly.HTMLElement) map[string]interface{}
}

// Service implements the Interface
type Service struct {
	Logger logger.Interface
}

// Ensure Service implements Interface
var _ Interface = (*Service)(nil)

// NewService creates a new Service instance
func NewService(logger logger.Interface) Interface {
	return &Service{Logger: logger}
}

type JSONLDMetadata struct {
	DateCreated    string                 `json:"dateCreated"`
	DateModified   string                 `json:"dateModified"`
	Description    string                 `json:"description"`
	Name           string                 `json:"name"`
	Type           string                 `json:"@type"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

// ExtractContent extracts content from an HTML element
func (s *Service) ExtractContent(e *colly.HTMLElement) *models.Content {
	var jsonLD JSONLDMetadata

	s.Logger.Debug("Extracting content", "url", e.Request.URL.String())

	// Extract metadata from JSON-LD first
	e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			s.Logger.Debug("Failed to parse JSON-LD", "error", err)
		}
	})

	// Extract metadata from meta tags and other sources
	metadata := s.ExtractMetadata(e)

	// Determine content type based on URL and page structure
	contentType := determineContentType(e.Request.URL.String(), metadata, jsonLD.Type)

	// Create content with basic fields
	content := &models.Content{
		ID:        uuid.New().String(),
		URL:       e.Request.URL.String(),
		Title:     getFirstNonEmpty(jsonLD.Name, e.ChildText("title"), e.ChildText("h1")),
		Body:      cleanBody(e),
		Type:      contentType,
		Metadata:  metadata,
		CreatedAt: parseDate([]string{jsonLD.DateCreated, jsonLD.DateModified}, s.Logger),
	}

	// Skip empty content
	if content.Title == "" && content.Body == "" {
		s.Logger.Debug("Skipping empty content", "url", content.URL)
		return nil
	}

	s.Logger.Debug("Extracted content",
		"id", content.ID,
		"title", content.Title,
		"url", content.URL,
		"type", content.Type,
		"created_at", content.CreatedAt)

	return content
}

// determineContentType determines the type of content based on URL and metadata
func determineContentType(url string, metadata map[string]interface{}, jsonLDType string) string {
	// Check JSON-LD type first
	if jsonLDType != "" {
		return jsonLDType
	}

	// Check metadata for content type
	if typeVal, ok := metadata["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			return typeStr
		}
	}

	// Check URL patterns
	switch {
	case strings.Contains(url, "/category/"), strings.Contains(url, "/categories/"):
		return "category"
	case strings.Contains(url, "/tag/"), strings.Contains(url, "/tags/"):
		return "tag"
	case strings.Contains(url, "/author/"), strings.Contains(url, "/authors/"):
		return "author"
	case strings.Contains(url, "/page/"), strings.Contains(url, "/pages/"):
		return "page"
	case strings.Contains(url, "/search"):
		return "search"
	case strings.Contains(url, "/feed"):
		return "feed"
	case strings.Contains(url, "/rss"):
		return "rss"
	case strings.Contains(url, "/sitemap"):
		return "sitemap"
	case strings.Contains(url, "/wp-json"), strings.Contains(url, "/wp-admin"):
		return "system"
	default:
		return "webpage"
	}
}

// ExtractMetadata extracts metadata from various sources in the HTML
func (s *Service) ExtractMetadata(e *colly.HTMLElement) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Extract OpenGraph metadata
	e.ForEach(`meta[property^="og:"]`, func(_ int, el *colly.HTMLElement) {
		property := el.Attr("property")
		content := el.Attr("content")
		if property != "" && content != "" {
			metadata[property[3:]] = content // Remove "og:" prefix
		}
	})

	// Extract Twitter metadata
	e.ForEach(`meta[name^="twitter:"]`, func(_ int, el *colly.HTMLElement) {
		name := el.Attr("name")
		content := el.Attr("content")
		if name != "" && content != "" {
			metadata[name[8:]] = content // Remove "twitter:" prefix
		}
	})

	// Extract other meta tags
	e.ForEach(`meta[name]`, func(_ int, el *colly.HTMLElement) {
		name := el.Attr("name")
		content := el.Attr("content")
		if name != "" && content != "" {
			metadata[name] = content
		}
	})

	return metadata
}

// Helper function to get the first non-empty string
func getFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Helper function to parse dates
func parseDate(dates []string, logger logger.Interface) time.Time {
	var parsedDate time.Time
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.2030000Z",
		"2006-01-02 15:04:05",
	}

	for _, dateStr := range dates {
		if dateStr == "" {
			continue
		}
		logger.Debug("Trying to parse date", "value", dateStr)
		for _, format := range timeFormats {
			t, err := time.Parse(format, dateStr)
			if err == nil {
				parsedDate = t
				logger.Debug("Successfully parsed date",
					"source", dateStr,
					"format", format,
					"result", t)
				break
			}
			logger.Debug("Failed to parse date",
				"source", dateStr,
				"format", format,
				"error", err)
		}
		if !parsedDate.IsZero() {
			break
		}
	}

	if parsedDate.IsZero() {
		logger.Debug("No valid date found", "dates", dates)
	}

	return parsedDate
}

// cleanBody removes excluded tags and elements from the body content
func cleanBody(e *colly.HTMLElement) string {
	// Get a copy of the body element for manipulation
	bodyEl := e.DOM

	// Remove all style and script tags directly
	bodyEl.Find("style,script").Remove()

	// Remove other excluded elements
	excludedSelectors := []string{
		"header", "footer", "nav", "aside",
		".advertisement", ".ads", ".social-share",
		".comments", ".related-posts", ".newsletter",
		"iframe", "noscript", "form", "button",
		".cookie-notice", ".popup", ".modal",
		".newsletter-signup", ".social-buttons",
	}

	for _, selector := range excludedSelectors {
		bodyEl.Find(selector).Remove()
	}

	// Get the cleaned text
	return bodyEl.Text()
}
