// Package processor provides content processing functionality for the application.
package processor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	// ErrInvalidHTML is returned when the HTML content is invalid.
	ErrInvalidHTML = errors.New("invalid HTML")
	// ErrMissingRequiredField is returned when a required field is missing.
	ErrMissingRequiredField = errors.New("missing required field")
)

// Process processes HTML content using the configured selectors.
func (p *HTMLProcessor) Process(html []byte) (*ProcessedContent, error) {
	// Check for invalid HTML first
	if len(html) == 0 || !bytes.Contains(html, []byte("<")) {
		return nil, fmt.Errorf("%w: not a valid HTML document", ErrInvalidHTML)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHTML, err)
	}

	content := Content{
		Title:       p.extractText(doc, p.Selectors["title"]),
		Body:        p.extractText(doc, p.Selectors["body"]),
		URL:         p.extractURL(doc),
		PublishedAt: p.extractTime(doc, p.Selectors["published_at"]),
		Author:      p.extractText(doc, p.Selectors["author"]),
		Categories:  p.extractList(doc, p.Selectors["categories"]),
		Tags:        p.extractList(doc, p.Selectors["tags"]),
		Metadata:    p.extractMetadata(doc),
	}

	// Only check for required fields if the HTML is valid
	if content.Title == "" {
		return nil, fmt.Errorf("%w: title", ErrMissingRequiredField)
	}

	return &ProcessedContent{
		Content: content,
	}, nil
}

// extractText extracts text from an element matching the selector.
func (p *HTMLProcessor) extractText(doc *goquery.Document, selector string) string {
	if selector == "" {
		return ""
	}
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		return
	})
	return strings.TrimSpace(doc.Find(selector).First().Text())
}

// extractURL extracts the URL from canonical link or og:url meta tag.
func (p *HTMLProcessor) extractURL(doc *goquery.Document) string {
	// Try canonical link
	if canonical := doc.Find("link[rel='canonical']").First(); canonical.Length() > 0 {
		if href, exists := canonical.Attr("href"); exists {
			return href
		}
	}

	// Try og:url meta tag
	if ogURL := doc.Find("meta[property='og:url']").First(); ogURL.Length() > 0 {
		if content, exists := ogURL.Attr("content"); exists {
			return content
		}
	}

	return ""
}

// extractTime extracts a time from an element matching the selector.
func (p *HTMLProcessor) extractTime(doc *goquery.Document, selector string) time.Time {
	if selector == "" {
		return time.Time{}
	}

	text := p.extractText(doc, selector)
	if text == "" {
		return time.Time{}
	}

	// Try various time formats
	formats := []string{
		time.RFC3339,
		time.RFC1123,
		time.RFC1123Z,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, text); err == nil {
			return t
		}
	}

	return time.Time{}
}

// extractList extracts a list of strings from an element matching the selector.
func (p *HTMLProcessor) extractList(doc *goquery.Document, selector string) []string {
	if selector == "" {
		return nil
	}

	var items []string
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			// Split on whitespace and filter out empty strings
			for _, item := range strings.Fields(text) {
				if item != "" {
					items = append(items, item)
				}
			}
		}
	})
	return items
}

// extractMetadata extracts metadata from meta tags.
func (p *HTMLProcessor) extractMetadata(doc *goquery.Document) map[string]string {
	metadata := make(map[string]string)

	// Extract meta tags
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				metadata[name] = content
			}
		}
		if property, exists := s.Attr("property"); exists {
			if content, exists := s.Attr("content"); exists {
				metadata[property] = content
			}
		}
	})

	return metadata
}
