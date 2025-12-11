// Package generator provides tools for generating CSS selector configurations
// for news sources.
package generator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	sampleTextLength        = 100
	longContentThreshold    = 500
	mediumContentThreshold  = 200
	highConfidenceThreshold = 0.95
	mediumHighConfidence    = 0.90
	mediumConfidence        = 0.85
	mediumLowConfidence     = 0.75
	lowConfidence           = 0.70
	veryLowConfidence       = 0.65
	minConfidence           = 0.60
	linkLowConfidence       = 0.70
	linkMinConfidence       = 0.75
)

// SelectorDiscovery analyzes HTML documents to discover CSS selectors
// for extracting article content.
type SelectorDiscovery struct {
	doc *goquery.Document
	url *url.URL
}

// NewSelectorDiscovery creates a new SelectorDiscovery instance.
func NewSelectorDiscovery(doc *goquery.Document, sourceURL string) (*SelectorDiscovery, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return &SelectorDiscovery{
		doc: doc,
		url: parsedURL,
	}, nil
}

// DiscoverAll runs all discovery methods and returns a complete DiscoveryResult.
func (sd *SelectorDiscovery) DiscoverAll() DiscoveryResult {
	return DiscoveryResult{
		Title:         sd.DiscoverTitle(),
		Body:          sd.DiscoverBody(),
		Author:        sd.DiscoverAuthor(),
		PublishedTime: sd.DiscoverPublishedTime(),
		Image:         sd.DiscoverImage(),
		Link:          sd.DiscoverLinks(),
		Category:      sd.DiscoverCategory(),
		Exclusions:    sd.DiscoverExclusions(),
	}
}

// DiscoverTitle finds title selectors with confidence scoring.
func (sd *SelectorDiscovery) DiscoverTitle() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "title",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Semantic HTML - highest confidence
	semanticSelectors := []string{"article h1", "main h1", "article h1.article-title", "main h1.page-title"}
	for _, sel := range semanticSelectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < highConfidenceThreshold {
			candidate.Confidence = highConfidenceThreshold
			candidate.SampleText = truncateText(text, sampleTextLength)
		}
	}

	// Schema.org and Open Graph - high confidence
	metaSelectors := []string{
		"meta[property='og:title']",
		"[itemprop='headline']",
		"meta[name='twitter:title']",
	}
	for _, sel := range metaSelectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < mediumHighConfidence {
			candidate.Confidence = mediumHighConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}

	// Class patterns - medium-high confidence
	classPatterns := []string{
		"h1[class*='title']",
		".article-title",
		".headline",
		".post-title",
		"h1.title",
		"h2.title",
	}
	for _, sel := range classPatterns {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		// Check for uniqueness - penalize if multiple matches
		count := sd.doc.Find(sel).Length()
		confidence := mediumLowConfidence
		if count > 1 {
			confidence = veryLowConfidence // Penalize ambiguity
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}

	// Fallback to any h1
	if len(candidate.Selectors) == 0 {
		text, found := sd.extractText("h1")
		if found && text != "" {
			count := sd.doc.Find("h1").Length()
			confidence := lowConfidence
			if count > 1 {
				confidence = minConfidence
			}
			candidate.Selectors = append(candidate.Selectors, "h1")
			candidate.Confidence = confidence
			candidate.SampleText = truncateText(text, sampleTextLength)
		}
	}

	return candidate
}

// DiscoverBody finds article body selectors.
func (sd *SelectorDiscovery) DiscoverBody() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "body",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Semantic HTML - highest confidence
	semanticSelectors := []string{
		"article",
		"[itemprop='articleBody']",
		"main article",
		"article .article-content",
	}
	const longContentThreshold = 500
	const mediumContentThreshold = 200
	for _, sel := range semanticSelectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		length := len(strings.TrimSpace(text))
		confidence := 0.90
		// Bonus for longer content (up to +0.1)
		if length > longContentThreshold {
			confidence = 0.95
		} else if length > mediumContentThreshold {
			confidence = 0.92
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			candidate.SampleText = truncateText(text, sampleTextLength)
		}
	}

	// Common class patterns
	classPatterns := []string{
		".article-body",
		".article-content",
		".post-content",
		".entry-content",
		".content",
		"main .content",
	}
	for _, sel := range classPatterns {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		length := len(strings.TrimSpace(text))
		confidence := 0.85
		// Bonus for longer content
		if length > longContentThreshold {
			confidence = 0.90
		} else if length > mediumContentThreshold {
			confidence = 0.87
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}

	return candidate
}

// DiscoverAuthor finds author selectors.
func (sd *SelectorDiscovery) DiscoverAuthor() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "author",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Schema.org and meta tags - highest confidence
	schemaSelectors := []string{
		"[itemprop='author']",
		"[rel='author']",
		"meta[property='article:author']",
		"meta[name='author']",
	}
	const authorMediumConfidence = 0.80
	for _, sel := range schemaSelectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < highConfidenceThreshold {
			candidate.Confidence = highConfidenceThreshold
			candidate.SampleText = truncateText(text, sampleTextLength)
		}
	}

	// Class patterns - medium confidence
	classPatterns := []string{
		".author",
		".byline",
		".article-author",
		".post-author",
		".writer",
	}
	for _, sel := range classPatterns {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < authorMediumConfidence {
			candidate.Confidence = authorMediumConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}

	return candidate
}

// DiscoverPublishedTime finds date/time selectors.
func (sd *SelectorDiscovery) DiscoverPublishedTime() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "published_time",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Meta tags - highest confidence
	metaSelectors := []string{
		"meta[property='article:published_time']",
		"meta[name='publishdate']",
		"meta[name='pubdate']",
		"meta[name='date']",
	}
	for _, sel := range metaSelectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < highConfidenceThreshold {
			candidate.Confidence = highConfidenceThreshold
			candidate.SampleText = truncateText(text, sampleTextLength)
		}
	}

	// Time element with datetime attribute
	timeText, timeFound := sd.extractText("time[datetime]")
	if timeFound && timeText != "" {
		candidate.Selectors = append(candidate.Selectors, "time[datetime]")
		if candidate.Confidence < mediumHighConfidence {
			candidate.Confidence = mediumHighConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(timeText, sampleTextLength)
			}
		}
	}

	// Schema.org
	schemaText, schemaFound := sd.extractText("[itemprop='datePublished']")
	if schemaFound && schemaText != "" {
		candidate.Selectors = append(candidate.Selectors, "[itemprop='datePublished']")
		if candidate.Confidence < highConfidenceThreshold {
			candidate.Confidence = highConfidenceThreshold
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(schemaText, sampleTextLength)
			}
		}
	}

	// Class patterns
	classPatterns := []string{
		".published-date",
		".date",
		".post-date",
		".article-date",
		".time",
		".timestamp",
	}
	for _, sel := range classPatterns {
		classText, classFound := sd.extractText(sel)
		if !classFound || classText == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < mediumLowConfidence {
			candidate.Confidence = mediumLowConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(classText, sampleTextLength)
			}
		}
	}

	return candidate
}

// DiscoverImage finds image selectors.
func (sd *SelectorDiscovery) DiscoverImage() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "image",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Open Graph - highest confidence
	if src, found := sd.extractAttr("meta[property='og:image']", "content"); found && src != "" {
		candidate.Selectors = append(candidate.Selectors, "meta[property='og:image']")
		candidate.Confidence = highConfidenceThreshold
		candidate.SampleText = truncateText(src, sampleTextLength)
	}

	// Schema.org
	if src, found := sd.extractAttr("[itemprop='image']", "src"); found && src != "" {
		candidate.Selectors = append(candidate.Selectors, "[itemprop='image']")
		if candidate.Confidence < mediumHighConfidence {
			candidate.Confidence = mediumHighConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(src, sampleTextLength)
			}
		}
	}

	// Article images
	articleImageSelectors := []string{
		"article img",
		"article picture img",
		".article-image img",
		".featured-image img",
		".post-image img",
	}
	for _, sel := range articleImageSelectors {
		src, found := sd.extractAttr(sel, "src")
		if !found || src == "" {
			continue
		}
		// Skip placeholder images
		if strings.Contains(src, "placeholder") || strings.Contains(src, "fallback") {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < mediumConfidence {
			candidate.Confidence = mediumConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(src, sampleTextLength)
			}
		}
	}

	return candidate
}

// DiscoverLinks finds article link patterns.
func (sd *SelectorDiscovery) DiscoverLinks() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "link",
		Selectors:  []string{},
		Confidence: linkLowConfidence, // Lower confidence - needs manual review
	}

	// Common article URL patterns
	patterns := []string{
		"/news/",
		"/article/",
		"/story/",
		"/post/",
		"/blog/",
		"/local-news/",
	}

	linkSelectors := make(map[string]int)
	var sampleHref string

	// Find all links and analyze their href patterns
	sd.doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Check if href matches any article pattern
		matched := false
		for _, pattern := range patterns {
			if strings.Contains(href, pattern) {
				matched = true
				if sampleHref == "" {
					sampleHref = href
				}
				break
			}
		}

		if matched {
			// Build a simple selector based on the link's attributes
			selector := sd.buildLinkSelector(s)
			if selector != "" {
				linkSelectors[selector]++
			}
		}
	})

	// Get top 5 most common selectors
	type selectorCount struct {
		selector string
		count    int
	}
	counts := make([]selectorCount, 0, len(linkSelectors))
	for sel, count := range linkSelectors {
		counts = append(counts, selectorCount{selector: sel, count: count})
	}

	// Sort by count (simple selection sort for top 5)
	for i := 0; i < len(counts) && i < 5; i++ {
		maxIdx := i
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[maxIdx].count {
				maxIdx = j
			}
		}
		counts[i], counts[maxIdx] = counts[maxIdx], counts[i]
		candidate.Selectors = append(candidate.Selectors, counts[i].selector)
	}

	// If no specific selectors found, use generic patterns
	if len(candidate.Selectors) == 0 {
		for _, pattern := range patterns {
			candidate.Selectors = append(candidate.Selectors, "a[href*='"+pattern+"']")
		}
	}

	if sampleHref != "" {
		candidate.SampleText = truncateText(sampleHref, sampleTextLength)
	}

	return candidate
}

// DiscoverCategory finds category selectors.
func (sd *SelectorDiscovery) DiscoverCategory() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "category",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Meta tags - high confidence
	text, found := sd.extractText("meta[property='article:section']")
	if found && text != "" {
		candidate.Selectors = append(candidate.Selectors, "meta[property='article:section']")
		candidate.Confidence = mediumHighConfidence
		candidate.SampleText = truncateText(text, sampleTextLength)
	}

	// Class patterns
	classPatterns := []string{
		".category",
		".section",
		".article-category",
		".post-category",
		"[data-category]",
	}
	for _, sel := range classPatterns {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}
		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < mediumLowConfidence {
			candidate.Confidence = mediumLowConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}

	return candidate
}

// DiscoverExclusions finds common elements to exclude.
func (sd *SelectorDiscovery) DiscoverExclusions() []string {
	exclusions := []string{}

	// Common exclusion patterns
	exclusionPatterns := []string{
		".ad",
		"[class*='ad__']",
		"[id^='ad-']",
		"[id*='ad__']",
		"[data-aqa='advertisement']",
		"[data-ad]",
		"nav",
		".header",
		".footer",
		"script",
		"style",
		"noscript",
		"[aria-hidden='true']",
		".visually-hidden",
		".social-follow",
		".share-buttons",
		"button",
		"form",
		".sidebar",
		".comments-section",
		".pagination",
		".related-posts",
		".newsletter-widget",
		".widget",
		".consent__banner",
		".cookie-banner",
	}

	for _, pattern := range exclusionPatterns {
		if sd.doc.Find(pattern).Length() > 0 {
			exclusions = append(exclusions, pattern)
		}
	}

	return exclusions
}

// Helper methods

// extractText extracts text content from a selector.
// For meta tags, it extracts the content attribute.
func (sd *SelectorDiscovery) extractText(selector string) (string, bool) {
	selection := sd.doc.Find(selector).First()
	if selection.Length() == 0 {
		return "", false
	}

	// Handle meta tags
	if strings.HasPrefix(selector, "meta[") {
		content, exists := selection.Attr("content")
		if exists {
			return strings.TrimSpace(content), true
		}
		return "", false
	}

	// Handle time elements with datetime
	if strings.Contains(selector, "time[datetime]") {
		datetime, exists := selection.Attr("datetime")
		if exists {
			return strings.TrimSpace(datetime), true
		}
	}

	// Regular text extraction
	text := strings.TrimSpace(selection.Text())
	if text == "" {
		return "", false
	}

	return text, true
}

// extractAttr extracts an attribute value from a selector.
func (sd *SelectorDiscovery) extractAttr(selector, attr string) (string, bool) {
	selection := sd.doc.Find(selector).First()
	if selection.Length() == 0 {
		return "", false
	}

	value, exists := selection.Attr(attr)
	if !exists {
		return "", false
	}

	return strings.TrimSpace(value), true
}

// buildLinkSelector builds a CSS selector for a link element.
func (sd *SelectorDiscovery) buildLinkSelector(s *goquery.Selection) string {
	tagName := goquery.NodeName(s)
	if tagName != "a" {
		return ""
	}

	// Check for ID
	if id, exists := s.Attr("id"); exists && id != "" {
		return "a#" + id
	}

	// Check for class
	if class, exists := s.Attr("class"); exists && class != "" {
		classes := strings.Fields(class)
		if len(classes) > 0 {
			// Use first class, but prefer article-related classes
			for _, c := range classes {
				if strings.Contains(c, "article") || strings.Contains(c, "link") || strings.Contains(c, "card") {
					return "a." + c
				}
			}
			return "a." + classes[0]
		}
	}

	// Check for data attributes
	if dataLink, exists := s.Attr("data-tb-link"); exists && dataLink != "" {
		return "a[data-tb-link]"
	}

	// Check href pattern
	if href, exists := s.Attr("href"); exists && href != "" {
		pattern := extractPattern(href)
		if pattern != "" {
			return "a[href*='" + pattern + "']"
		}
	}

	return ""
}

// extractPattern extracts a URL pattern from a href.
func extractPattern(href string) string {
	patterns := []string{"/news/", "/article/", "/story/", "/post/", "/blog/", "/local-news/"}
	for _, pattern := range patterns {
		if strings.Contains(href, pattern) {
			return pattern
		}
	}
	return ""
}

// truncateText truncates text to a maximum length.
// maxLen is always 100, but kept as parameter for API consistency.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
