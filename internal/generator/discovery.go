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
	// Body selector confidence values
	semanticBodyConfidence = 0.90
	classBodyConfidence    = 0.85
	// Confidence bonuses for body content length
	longContentBonus   = 0.05
	mediumContentBonus = 0.02
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

	semanticSelectors := []string{"article h1", "main h1", "article h1.article-title", "main h1.page-title"}
	sd.processTitleSelectors(&candidate, semanticSelectors, highConfidenceThreshold, false)

	metaSelectors := []string{
		"meta[property='og:title']",
		"[itemprop='headline']",
		"meta[name='twitter:title']",
	}
	sd.processTitleSelectors(&candidate, metaSelectors, mediumHighConfidence, false)

	classPatterns := []string{
		"h1[class*='title']",
		".article-title",
		".headline",
		".post-title",
		"h1.title",
		"h2.title",
	}
	sd.processTitleSelectorsWithUniqueness(&candidate, classPatterns)

	sd.fallbackToH1(&candidate)

	return candidate
}

// processTitleSelectors processes title selectors with a fixed confidence.
func (sd *SelectorDiscovery) processTitleSelectors(
	candidate *SelectorCandidate,
	selectors []string,
	confidence float64,
	checkUniqueness bool,
) {
	for _, sel := range selectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}

		actualConfidence := confidence
		if checkUniqueness {
			count := sd.doc.Find(sel).Length()
			if count > 1 {
				actualConfidence = veryLowConfidence
			}
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < actualConfidence {
			candidate.Confidence = actualConfidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}
}

// processTitleSelectorsWithUniqueness processes selectors with uniqueness checking.
func (sd *SelectorDiscovery) processTitleSelectorsWithUniqueness(
	candidate *SelectorCandidate,
	selectors []string,
) {
	for _, sel := range selectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}

		count := sd.doc.Find(sel).Length()
		confidence := mediumLowConfidence
		if count > 1 {
			confidence = veryLowConfidence
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}
}

// fallbackToH1 falls back to any h1 if no selectors found.
func (sd *SelectorDiscovery) fallbackToH1(candidate *SelectorCandidate) {
	if len(candidate.Selectors) > 0 {
		return
	}

	text, found := sd.extractText("h1")
	if !found || text == "" {
		return
	}

	count := sd.doc.Find("h1").Length()
	confidence := lowConfidence
	if count > 1 {
		confidence = minConfidence
	}

	candidate.Selectors = append(candidate.Selectors, "h1")
	candidate.Confidence = confidence
	candidate.SampleText = truncateText(text, sampleTextLength)
}

// DiscoverBody finds article body selectors.
func (sd *SelectorDiscovery) DiscoverBody() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "body",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	semanticSelectors := []string{
		"article",
		"[itemprop='articleBody']",
		"main article",
		"article .article-content",
	}
	sd.processBodySelectors(&candidate, semanticSelectors, semanticBodyConfidence)

	classPatterns := []string{
		".article-body",
		".article-content",
		".post-content",
		".entry-content",
		".content",
		"main .content",
	}
	sd.processBodySelectors(&candidate, classPatterns, classBodyConfidence)

	return candidate
}

// processBodySelectors processes body selectors with confidence calculation.
func (sd *SelectorDiscovery) processBodySelectors(
	candidate *SelectorCandidate,
	selectors []string,
	baseConfidence float64,
) {
	const longContentThreshold = 500
	const mediumContentThreshold = 200

	for _, sel := range selectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}

		length := len(strings.TrimSpace(text))
		confidence := calculateBodyConfidence(length, baseConfidence, longContentThreshold, mediumContentThreshold)

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}
}

// calculateBodyConfidence calculates confidence based on content length.
func calculateBodyConfidence(length int, baseConfidence, longThreshold, mediumThreshold float64) float64 {
	if length > int(longThreshold) {
		return baseConfidence + longContentBonus
	}
	if length > int(mediumThreshold) {
		return baseConfidence + mediumContentBonus
	}
	return baseConfidence
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

	metaSelectors := []string{
		"meta[property='article:published_time']",
		"meta[name='publishdate']",
		"meta[name='pubdate']",
		"meta[name='date']",
	}
	sd.processPublishedTimeSelectors(&candidate, metaSelectors, highConfidenceThreshold)

	sd.checkTimeElement(&candidate)
	sd.checkSchemaOrgDate(&candidate)

	classPatterns := []string{
		".published-date",
		".date",
		".post-date",
		".article-date",
		".time",
		".timestamp",
	}
	sd.processPublishedTimeSelectors(&candidate, classPatterns, mediumLowConfidence)

	return candidate
}

// processPublishedTimeSelectors processes published time selectors.
func (sd *SelectorDiscovery) processPublishedTimeSelectors(
	candidate *SelectorCandidate,
	selectors []string,
	confidence float64,
) {
	for _, sel := range selectors {
		text, found := sd.extractText(sel)
		if !found || text == "" {
			continue
		}

		candidate.Selectors = append(candidate.Selectors, sel)
		if candidate.Confidence < confidence {
			candidate.Confidence = confidence
			if candidate.SampleText == "" {
				candidate.SampleText = truncateText(text, sampleTextLength)
			}
		}
	}
}

// checkTimeElement checks for time element with datetime.
func (sd *SelectorDiscovery) checkTimeElement(candidate *SelectorCandidate) {
	timeText, timeFound := sd.extractText("time[datetime]")
	if !timeFound || timeText == "" {
		return
	}

	candidate.Selectors = append(candidate.Selectors, "time[datetime]")
	if candidate.Confidence < mediumHighConfidence {
		candidate.Confidence = mediumHighConfidence
		if candidate.SampleText == "" {
			candidate.SampleText = truncateText(timeText, sampleTextLength)
		}
	}
}

// checkSchemaOrgDate checks for Schema.org datePublished.
func (sd *SelectorDiscovery) checkSchemaOrgDate(candidate *SelectorCandidate) {
	schemaText, schemaFound := sd.extractText("[itemprop='datePublished']")
	if !schemaFound || schemaText == "" {
		return
	}

	candidate.Selectors = append(candidate.Selectors, "[itemprop='datePublished']")
	if candidate.Confidence < highConfidenceThreshold {
		candidate.Confidence = highConfidenceThreshold
		if candidate.SampleText == "" {
			candidate.SampleText = truncateText(schemaText, sampleTextLength)
		}
	}
}

// DiscoverImage finds image selectors.
func (sd *SelectorDiscovery) DiscoverImage() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "image",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	sd.checkOpenGraphImage(&candidate)
	sd.checkSchemaOrgImage(&candidate)
	sd.checkArticleImages(&candidate)

	return candidate
}

// checkOpenGraphImage checks for Open Graph image.
func (sd *SelectorDiscovery) checkOpenGraphImage(candidate *SelectorCandidate) {
	src, found := sd.extractAttr("meta[property='og:image']", "content")
	if !found || src == "" {
		return
	}

	candidate.Selectors = append(candidate.Selectors, "meta[property='og:image']")
	candidate.Confidence = highConfidenceThreshold
	candidate.SampleText = truncateText(src, sampleTextLength)
}

// checkSchemaOrgImage checks for Schema.org image.
func (sd *SelectorDiscovery) checkSchemaOrgImage(candidate *SelectorCandidate) {
	src, found := sd.extractAttr("[itemprop='image']", "src")
	if !found || src == "" {
		return
	}

	candidate.Selectors = append(candidate.Selectors, "[itemprop='image']")
	if candidate.Confidence < mediumHighConfidence {
		candidate.Confidence = mediumHighConfidence
		if candidate.SampleText == "" {
			candidate.SampleText = truncateText(src, sampleTextLength)
		}
	}
}

// checkArticleImages checks for article image selectors.
func (sd *SelectorDiscovery) checkArticleImages(candidate *SelectorCandidate) {
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

		if isPlaceholderImage(src) {
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
}

// isPlaceholderImage checks if an image URL is a placeholder.
func isPlaceholderImage(src string) bool {
	return strings.Contains(src, "placeholder") || strings.Contains(src, "fallback")
}

// DiscoverLinks finds article link patterns.
func (sd *SelectorDiscovery) DiscoverLinks() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "link",
		Selectors:  []string{},
		Confidence: linkLowConfidence,
	}

	patterns := []string{
		"/news/",
		"/article/",
		"/story/",
		"/post/",
		"/blog/",
		"/local-news/",
	}

	linkSelectors, sampleHref := sd.collectLinkSelectors(patterns)
	candidate.Selectors = sd.getTopLinkSelectors(linkSelectors)

	if len(candidate.Selectors) == 0 {
		candidate.Selectors = sd.getGenericLinkSelectors(patterns)
	}

	if sampleHref != "" {
		candidate.SampleText = truncateText(sampleHref, sampleTextLength)
	}

	return candidate
}

// collectLinkSelectors collects link selectors from the document.
func (sd *SelectorDiscovery) collectLinkSelectors(patterns []string) (linkSelectors map[string]int, sampleHref string) {
	linkSelectors = make(map[string]int)

	sd.doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		if !matchesArticlePattern(href, patterns) {
			return
		}

		if sampleHref == "" {
			sampleHref = href
		}

		selector := sd.buildLinkSelector(s)
		if selector != "" {
			linkSelectors[selector]++
		}
	})

	return linkSelectors, sampleHref
}

// matchesArticlePattern checks if href matches any article pattern.
func matchesArticlePattern(href string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(href, pattern) {
			return true
		}
	}
	return false
}

// getTopLinkSelectors gets the top 5 most common selectors.
func (sd *SelectorDiscovery) getTopLinkSelectors(linkSelectors map[string]int) []string {
	type selectorCount struct {
		selector string
		count    int
	}

	counts := make([]selectorCount, 0, len(linkSelectors))
	for sel, count := range linkSelectors {
		counts = append(counts, selectorCount{selector: sel, count: count})
	}

	// Simple selection sort for top 5
	const topN = 5
	for i := 0; i < len(counts) && i < topN; i++ {
		maxIdx := i
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[maxIdx].count {
				maxIdx = j
			}
		}
		counts[i], counts[maxIdx] = counts[maxIdx], counts[i]
	}

	result := make([]string, 0, topN)
	for i := 0; i < len(counts) && i < topN; i++ {
		result = append(result, counts[i].selector)
	}

	return result
}

// getGenericLinkSelectors returns generic link selectors based on patterns.
func (sd *SelectorDiscovery) getGenericLinkSelectors(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		result = append(result, "a[href*='"+pattern+"']")
	}
	return result
}

// DiscoverCategory finds category selectors.
func (sd *SelectorDiscovery) DiscoverCategory() SelectorCandidate {
	candidate := SelectorCandidate{
		Field:      "category",
		Selectors:  []string{},
		Confidence: 0.0,
	}

	// Meta tags - high confidence
	categoryText, categoryFound := sd.extractText("meta[property='article:section']")
	if categoryFound && categoryText != "" {
		candidate.Selectors = append(candidate.Selectors, "meta[property='article:section']")
		candidate.Confidence = mediumHighConfidence
		candidate.SampleText = truncateText(categoryText, sampleTextLength)
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
	if goquery.NodeName(s) != "a" {
		return ""
	}

	if selector := buildLinkSelectorFromID(s); selector != "" {
		return selector
	}

	if selector := buildLinkSelectorFromClass(s); selector != "" {
		return selector
	}

	if selector := buildLinkSelectorFromDataAttr(s); selector != "" {
		return selector
	}

	return buildLinkSelectorFromHref(s)
}

// buildLinkSelectorFromID builds selector from ID attribute.
func buildLinkSelectorFromID(s *goquery.Selection) string {
	id, exists := s.Attr("id")
	if exists && id != "" {
		return "a#" + id
	}
	return ""
}

// buildLinkSelectorFromClass builds selector from class attribute.
func buildLinkSelectorFromClass(s *goquery.Selection) string {
	class, exists := s.Attr("class")
	if !exists || class == "" {
		return ""
	}

	classes := strings.Fields(class)
	if len(classes) == 0 {
		return ""
	}

	for _, c := range classes {
		if isArticleRelatedClass(c) {
			return "a." + c
		}
	}

	return "a." + classes[0]
}

// isArticleRelatedClass checks if a class name is article-related.
func isArticleRelatedClass(className string) bool {
	return strings.Contains(className, "article") ||
		strings.Contains(className, "link") ||
		strings.Contains(className, "card")
}

// buildLinkSelectorFromDataAttr builds selector from data attribute.
func buildLinkSelectorFromDataAttr(s *goquery.Selection) string {
	dataLink, exists := s.Attr("data-tb-link")
	if exists && dataLink != "" {
		return "a[data-tb-link]"
	}
	return ""
}

// buildLinkSelectorFromHref builds selector from href pattern.
func buildLinkSelectorFromHref(s *goquery.Selection) string {
	href, exists := s.Attr("href")
	if !exists || href == "" {
		return ""
	}

	pattern := extractPattern(href)
	if pattern == "" {
		return ""
	}

	return "a[href*='" + pattern + "']"
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
// maxLen is always sampleTextLength (100), but kept as parameter for API consistency.
func truncateText(text string, maxLen int) string { //nolint:unparam // kept for API consistency
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
