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
	start := time.Now()
	defer func() {
		p.metrics.RecordProcessingTime(time.Since(start))
	}()

	// Check for invalid HTML first
	if len(html) == 0 || !bytes.Contains(html, []byte("<")) {
		p.metrics.RecordError()
		return nil, fmt.Errorf("%w: not a valid HTML document", ErrInvalidHTML)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		p.metrics.RecordError()
		return nil, fmt.Errorf("%w: parse error", ErrInvalidHTML)
	}

	content := Content{
		Title:       p.extractText(doc, p.selectors["title"]),
		Body:        p.extractText(doc, p.selectors["body"]),
		URL:         p.extractURL(doc),
		PublishedAt: p.extractTime(doc, p.selectors["published_at"]),
		Author:      p.extractText(doc, p.selectors["author"]),
		Categories:  p.extractList(doc, p.selectors["categories"]),
		Tags:        p.extractList(doc, p.selectors["tags"]),
		Metadata:    p.extractMetadata(doc),
	}

	// Only check for required fields if we're processing the full content
	if p.selectors["title"] != "" && content.Title == "" {
		p.metrics.RecordError()
		return nil, fmt.Errorf("%w: title", ErrMissingRequiredField)
	}

	p.metrics.RecordElementsProcessed(1)
	return &ProcessedContent{
		Content: content,
	}, nil
}

// extractText extracts text from an element matching the selector.
func (p *HTMLProcessor) extractText(doc *goquery.Document, selector string) string {
	if selector == "" {
		return ""
	}
	text := strings.TrimSpace(doc.Find(selector).First().Text())
	if text != "" {
		p.metrics.RecordElementsProcessed(1)
	}
	return text
}

// extractURL extracts the URL from canonical link or og:url meta tag.
func (p *HTMLProcessor) extractURL(doc *goquery.Document) string {
	// Try canonical link
	if canonical := doc.Find("link[rel='canonical']").First(); canonical.Length() > 0 {
		if href, exists := canonical.Attr("href"); exists {
			p.metrics.RecordElementsProcessed(1)
			return href
		}
	}

	// Try og:url meta tag
	if ogURL := doc.Find("meta[property='og:url']").First(); ogURL.Length() > 0 {
		if content, exists := ogURL.Attr("content"); exists {
			p.metrics.RecordElementsProcessed(1)
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
			p.metrics.RecordElementsProcessed(1)
			return t
		}
	}

	p.metrics.RecordError()
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
					p.metrics.RecordElementsProcessed(1)
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
		if name, nameOK := s.Attr("name"); nameOK {
			if content, contentOK := s.Attr("content"); contentOK {
				metadata[name] = content
				p.metrics.RecordElementsProcessed(1)
			}
		}
		if property, propertyOK := s.Attr("property"); propertyOK {
			if content, contentOK := s.Attr("content"); contentOK {
				metadata[property] = content
				p.metrics.RecordElementsProcessed(1)
			}
		}
	})

	return metadata
}
