// Package content handles non-article content operations
package content

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the interface for content operations
type Interface interface {
	ExtractContent(e *colly.HTMLElement) *models.Content
	ExtractMetadata(e *colly.HTMLElement) map[string]any
	DetermineContentType(url string, metadata map[string]any, jsonLDType string) string
	Process(ctx context.Context, input string) string
	ProcessBatch(ctx context.Context, input []string) []string
	ProcessWithMetadata(ctx context.Context, input string, metadata map[string]string) string
}

// Service implements the Interface
type Service struct {
	Logger    logger.Interface
	Storage   types.Interface
	IndexName string
	// SourceRules maps source names to their content processing rules
	SourceRules map[string]ContentRules
	// DefaultRules is used when no source-specific rules are found
	DefaultRules ContentRules
}

// ContentRules defines the rules for processing content from a specific source
type ContentRules struct {
	// ContentTypePatterns maps content types to their URL patterns
	ContentTypePatterns map[string][]string
	// ExcludePatterns are URL patterns to exclude from processing
	ExcludePatterns []string
	// MetadataSelectors defines selectors for extracting metadata
	MetadataSelectors map[string]string
	// ContentSelectors defines selectors for extracting content
	ContentSelectors map[string]string
}

// Ensure Service implements Interface
var _ Interface = (*Service)(nil)

// NewService creates a new Service instance
func NewService(logger logger.Interface, storage types.Interface) Interface {
	return &Service{
		Logger:      logger,
		Storage:     storage,
		SourceRules: make(map[string]ContentRules),
		DefaultRules: ContentRules{
			ContentTypePatterns: contentTypePatterns,
			MetadataSelectors: map[string]string{
				"title":       "title",
				"description": "meta[name=description]",
				"keywords":    "meta[name=keywords]",
				"author":      "meta[name=author]",
			},
			ContentSelectors: map[string]string{
				"body": "article",
			},
		},
	}
}

// AddSourceRules adds content processing rules for a specific source
func (s *Service) AddSourceRules(sourceName string, rules ContentRules) {
	s.SourceRules[sourceName] = rules
}

// getRulesForURL returns the appropriate content rules for the given URL
func (s *Service) getRulesForURL(url string) ContentRules {
	// Try to find matching source rules
	for sourceName, rules := range s.SourceRules {
		if strings.Contains(url, sourceName) {
			return rules
		}
	}
	return s.DefaultRules
}

type JSONLDMetadata struct {
	DateCreated    string         `json:"dateCreated"`
	DateModified   string         `json:"dateModified"`
	Description    string         `json:"description"`
	Name           string         `json:"name"`
	Type           string         `json:"@type"`
	AdditionalData map[string]any `json:"additionalData,omitempty"`
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.2030000Z",
	"2006-01-02 15:04:05",
}

var contentTypePatterns = map[string][]string{
	"category": {"/category/", "/categories/"},
	"tag":      {"/tag/", "/tags/"},
	"author":   {"/author/", "/authors/"},
	"page":     {"/page/", "/pages/"},
	"search":   {"/search"},
	"feed":     {"/feed"},
	"rss":      {"/rss"},
	"sitemap":  {"/sitemap"},
	"system":   {"/wp-json", "/wp-admin"},
}

// cleanBody cleans and extracts text from HTML content
func cleanBody(content any) string {
	var selection *goquery.Selection

	switch v := content.(type) {
	case *colly.HTMLElement:
		selection = v.DOM
	case *goquery.Selection:
		selection = v
	default:
		return ""
	}

	// Remove unwanted elements
	selection.Find("script, style, noscript, iframe, form").Remove()
	return strings.TrimSpace(selection.Text())
}

// ExtractContent extracts content from an HTML element
func (s *Service) ExtractContent(e *colly.HTMLElement) *models.Content {
	s.Logger.Debug("Extracting content", "url", e.Request.URL.String())

	// Get rules for this URL
	rules := s.getRulesForURL(e.Request.URL.String())

	// Check if URL matches any exclude patterns
	url := e.Request.URL.String()
	for _, pattern := range rules.ExcludePatterns {
		if strings.Contains(url, pattern) {
			s.Logger.Debug("URL matches exclude pattern, skipping",
				"url", url,
				"pattern", pattern)
			return nil
		}
	}

	var jsonLD JSONLDMetadata
	var parsedDate time.Time

	// Extract metadata
	metadata := s.ExtractMetadata(e)

	// Parse JSON-LD if available
	if jsonLDStr := e.DOM.Find("script[type='application/ld+json']").Text(); jsonLDStr != "" {
		if err := json.Unmarshal([]byte(jsonLDStr), &jsonLD); err != nil {
			s.Logger.Error("Failed to parse JSON-LD", "error", err)
		}
	}

	// Try to parse date from various sources
	dates := []string{
		jsonLD.DateCreated,
		jsonLD.DateModified,
	}

	// Add metadata dates if they exist
	if publishedTime, ok := metadata["published_time"].(string); ok {
		dates = append(dates, publishedTime)
	}
	if modifiedTime, ok := metadata["modified_time"].(string); ok {
		dates = append(dates, modifiedTime)
	}

	for _, dateStr := range dates {
		if dateStr == "" {
			continue
		}
		parsedDate = s.parseDate(s.Logger, dateStr)
		if !parsedDate.IsZero() {
			break
		}
	}

	// Determine content type using source-specific patterns
	contentType := s.DetermineContentType(e.Request.URL.String(), metadata, jsonLD.Type)

	// Extract body using source-specific selectors
	body := ""
	if bodySelector, ok := rules.ContentSelectors["body"]; ok && bodySelector != "" {
		body = cleanBody(e.DOM.Find(bodySelector))
	} else {
		body = cleanBody(e)
	}

	// Create content object
	content := &models.Content{
		ID:        uuid.New().String(),
		Title:     jsonLD.Name,
		URL:       e.Request.URL.String(),
		Type:      contentType,
		Body:      body,
		CreatedAt: parsedDate,
		Metadata:  metadata,
	}

	s.Logger.Debug("Extracted content",
		"id", content.ID,
		"title", content.Title,
		"url", content.URL,
		"type", content.Type,
		"created_at", content.CreatedAt,
	)

	return content
}

// DetermineContentType determines the content type based on URL and metadata
func (s *Service) DetermineContentType(url string, metadata map[string]any, jsonLDType string) string {
	// Get rules for this URL
	rules := s.getRulesForURL(url)

	// First try JSON-LD type
	if jsonLDType != "" {
		return strings.ToLower(jsonLDType)
	}

	// Then try metadata type
	if metaType, ok := metadata["type"].(string); ok {
		return strings.ToLower(metaType)
	}

	// Try to detect from URL using source-specific patterns
	urlLower := strings.ToLower(url)
	for category, patterns := range rules.ContentTypePatterns {
		for _, pattern := range patterns {
			if strings.Contains(urlLower, pattern) {
				return category
			}
		}
	}

	// Default to webpage
	return "webpage"
}

// ExtractMetadata extracts metadata from various sources in the HTML
func (s *Service) ExtractMetadata(e *colly.HTMLElement) map[string]any {
	// Get rules for this URL
	rules := s.getRulesForURL(e.Request.URL.String())
	metadata := make(map[string]any)

	// Extract metadata using source-specific selectors
	for key, selector := range rules.MetadataSelectors {
		if value := e.ChildAttr(selector, "content"); value != "" {
			metadata[key] = value
		} else if value := e.ChildText(selector); value != "" {
			metadata[key] = value
		}
	}

	// Extract OpenGraph metadata (highest precedence)
	e.ForEach(`meta[property^="og:"]`, func(_ int, el *colly.HTMLElement) {
		property := el.Attr("property")
		content := el.Attr("content")
		if property != "" && content != "" {
			key := property[3:] // Remove "og:" prefix
			if _, exists := metadata[key]; !exists {
				metadata[key] = content
			}
		}
	})

	// Extract Twitter metadata (second precedence)
	e.ForEach(`meta[name^="twitter:"]`, func(_ int, el *colly.HTMLElement) {
		name := el.Attr("name")
		content := el.Attr("content")
		if name != "" && content != "" {
			key := name[8:] // Remove "twitter:" prefix
			if _, exists := metadata[key]; !exists {
				metadata[key] = content
			}
		}
	})

	return metadata
}

// parseDate attempts to parse a date string using multiple formats
func (s *Service) parseDate(logger logger.Interface, dateStr string) time.Time {
	logger.Debug("Trying to parse date", "value", dateStr)

	for _, format := range timeFormats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			logger.Debug("Successfully parsed date",
				"source", dateStr,
				"format", format,
				"result", t.String(),
			)
			return t
		}
		logger.Debug("Failed to parse date",
			"source", dateStr,
			"format", format,
			"error", err.Error(),
		)
	}

	return time.Time{}
}

// Process processes a single string content
func (s *Service) Process(_ context.Context, input string) string {
	s.Logger.Debug("Processing content", "input", input)

	// Create a reader from the input string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		s.Logger.Error("Failed to parse HTML", "error", err)
		return strings.TrimSpace(input)
	}

	// Get text content without HTML tags
	text := doc.Text()

	// Clean up whitespace
	text = strings.Join(strings.Fields(text), " ")

	s.Logger.Debug("Processed content", "result", text)
	return text
}

// ProcessBatch processes a batch of strings
func (s *Service) ProcessBatch(ctx context.Context, input []string) []string {
	result := make([]string, len(input))
	for i, str := range input {
		result[i] = s.Process(ctx, str)
	}
	return result
}

// ProcessWithMetadata processes content with additional metadata
func (s *Service) ProcessWithMetadata(ctx context.Context, input string, metadata map[string]string) string {
	if metadata != nil {
		s.Logger.Debug("Processing content with metadata",
			"source", metadata["source"],
			"type", metadata["type"],
		)
	}
	return s.Process(ctx, input)
}
