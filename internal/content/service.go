// Package content handles non-article content operations
package content

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the interface for content operations
type Interface interface {
	ExtractContent(e *colly.HTMLElement) *models.Content
	ExtractMetadata(e *colly.HTMLElement) map[string]any
	DetermineContentType(url string, metadata map[string]any, jsonLDType string) common.ContentType
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
	ContentTypePatterns map[common.ContentType][]string
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

var contentTypePatterns = map[common.ContentType][]string{
	common.ContentTypeArticle: {"/article/", "/articles/", "/post/", "/posts/"},
	common.ContentTypePage:    {"/page/", "/pages/"},
	common.ContentTypeVideo:   {"/video/", "/videos/"},
	common.ContentTypeImage:   {"/image/", "/images/", "/photo/", "/photos/", "/gallery/"},
	common.ContentTypeHTML:    {"/html/", "/htm/"},
	common.ContentTypeJob:     {"/job/", "/jobs/", "/career/", "/careers/"},
}

// ExtractContent extracts content from an HTML element
func (s *Service) ExtractContent(e *colly.HTMLElement) *models.Content {
	content := &models.Content{
		URL:  e.Request.URL.String(),
		Body: strings.TrimSpace(e.Text),
	}

	return content
}

// DetermineContentType determines the content type based on URL and metadata
func (s *Service) DetermineContentType(url string, metadata map[string]any, jsonLDType string) common.ContentType {
	// Get rules for this URL
	rules := s.getRulesForURL(url)

	// First try JSON-LD type
	if jsonLDType != "" {
		s.Logger.Debug("Using JSON-LD type",
			"url", url,
			"type", jsonLDType)
		return common.ContentType(strings.ToLower(jsonLDType))
	}

	// Then try metadata type
	if metaType, ok := metadata["type"].(string); ok {
		s.Logger.Debug("Using metadata type",
			"url", url,
			"type", metaType)
		return common.ContentType(strings.ToLower(metaType))
	}

	// Try to detect from URL using source-specific patterns
	urlLower := strings.ToLower(url)
	for contentType, patterns := range rules.ContentTypePatterns {
		for _, pattern := range patterns {
			if strings.Contains(urlLower, pattern) {
				s.Logger.Debug("Using URL pattern type",
					"url", url,
					"pattern", pattern,
					"type", contentType)
				return contentType
			}
		}
	}

	// Default to webpage
	s.Logger.Debug("Using default type",
		"url", url,
		"type", common.ContentTypePage)
	return common.ContentTypePage
}

// ExtractMetadata extracts metadata from an HTML element
func (s *Service) ExtractMetadata(e *colly.HTMLElement) map[string]any {
	// Get rules for this URL
	rules := s.getRulesForURL(e.Request.URL.String())
	metadata := make(map[string]any)

	// Extract metadata using source-specific selectors
	for key, selector := range rules.MetadataSelectors {
		if value := e.DOM.Find(selector).AttrOr("content", ""); value != "" {
			metadata[key] = value
		} else if textValue := e.DOM.Find(selector).Text(); textValue != "" {
			metadata[key] = textValue
		}
	}

	// Extract OpenGraph metadata (highest precedence)
	e.DOM.Find(`meta[property^="og:"], meta[property^="article:"]`).Each(func(_ int, s *goquery.Selection) {
		property, exists := s.Attr("property")
		if !exists {
			return
		}
		content, exists := s.Attr("content")
		if !exists {
			return
		}
		metadata[property] = content
	})

	// Extract Twitter metadata (second precedence)
	e.DOM.Find(`meta[name^="twitter:"]`).Each(func(_ int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if !exists {
			return
		}
		content, exists := s.Attr("content")
		if !exists {
			return
		}
		key := name[8:] // Remove "twitter:" prefix
		if _, exists := metadata[key]; !exists {
			metadata[key] = content
		}
	})

	s.Logger.Debug("Extracted metadata",
		"url", e.Request.URL.String(),
		"metadata", metadata)

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
func (s *Service) Process(ctx context.Context, html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		s.Logger.Error("failed to parse HTML", err)
		return ""
	}

	var text []string
	doc.Find("p, div").Each(func(i int, sel *goquery.Selection) {
		t := strings.TrimSpace(sel.Text())
		if t != "" {
			text = append(text, t)
		}
	})

	return strings.Join(text, " ")
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

// ExtractText extracts the text content from a goquery Document.
func (s *Service) ExtractText(doc *goquery.Document) string {
	return doc.Find("body").Text()
}

func (s *Service) ExtractContentFromDocument(url string, doc *goquery.Document) (*models.Content, error) {
	body := s.ExtractText(doc)
	if body == "" {
		return nil, fmt.Errorf("no content found in document")
	}

	return &models.Content{
		URL:       url,
		Body:      body,
		CreatedAt: time.Now(),
	}, nil
}
