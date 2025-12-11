// Package articles provides functionality for processing and managing article content.
package articles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract basic fields (before applying excludes, as these are usually in head or specific locations)
	extractBasicFields(data, e, selectors)

	// Extract body content
	extractBodyContent(data, e, selectors)

	// Extract metadata
	extractMetadata(data, e, selectors)

	// Extract tags
	extractTags(data, e, selectors)

	// Extract Open Graph metadata
	extractOpenGraphMetadata(data, e)

	// Extract other metadata
	extractOtherMetadata(data, e, selectors, sourceURL)

	// Extract article ID
	data.ID = extractArticleID(e, selectors, sourceURL)

	return data
}

// extractBasicFields extracts title and intro fields.
func extractBasicFields(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
	// Extract title from selector
	extractedTitle := extractText(e, selectors.Title)

	// Extract OG title as fallback/preference
	ogTitle := extractMeta(e, "og:title")

	// Prefer OG title if available, otherwise use extracted title
	if ogTitle != "" {
		data.Title = ogTitle
	} else {
		data.Title = extractedTitle
	}

	data.Intro = extractText(e, selectors.Intro)
	if data.Intro == "" {
		// Fallback to OG description
		data.Intro = extractMeta(e, "og:description")
	}
}

// extractBodyContent extracts the article body content.
func extractBodyContent(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
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
}

// extractMetadata extracts author, byline, and published date.
func extractMetadata(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
	data.Author = extractText(e, selectors.Author)
	if data.Author == "" {
		data.Author = extractMeta(e, "article:author")
	}

	data.BylineName = extractText(e, selectors.BylineName)
	if data.BylineName == "" {
		data.BylineName = extractText(e, selectors.Byline)
	}

	// Extract dates with multiple fallback strategies
	data.PublishedDate = extractPublishedDate(e, selectors)
}

// extractTags extracts tags and keywords from the article.
func extractTags(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
	extractKeywords(data, e, selectors)
	extractTagsFromSelector(data, e, selectors)
}

// extractKeywords extracts keywords and adds them to data.
func extractKeywords(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
	keywordsStr := extractText(e, selectors.Keywords)
	if keywordsStr == "" {
		keywordsStr = extractMetaName(e, "keywords")
	}
	if keywordsStr == "" {
		return
	}

	keywords := parseCommaSeparatedList(keywordsStr)
	data.Keywords = append(data.Keywords, keywords...)
	// Also add to tags for backward compatibility
	data.Tags = append(data.Tags, keywords...)
}

// extractTagsFromSelector extracts tags from the tags selector.
func extractTagsFromSelector(data *ArticleData, e *colly.HTMLElement, selectors configtypes.ArticleSelectors) {
	tagsStr := extractText(e, selectors.Tags)
	if tagsStr == "" {
		return
	}

	tags := parseCommaSeparatedList(tagsStr)
	keywordSet := makeStringSet(data.Keywords)
	tagSet := makeStringSet(data.Tags)

	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if !keywordSet[tag] && !tagSet[tag] {
			data.Tags = append(data.Tags, tag)
			tagSet[tag] = true
		}
	}
}

// parseCommaSeparatedList parses a comma-separated string into a slice.
func parseCommaSeparatedList(s string) []string {
	items := strings.Split(s, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

// makeStringSet creates a map for fast lookup.
func makeStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

// extractOpenGraphMetadata extracts Open Graph metadata.
func extractOpenGraphMetadata(data *ArticleData, e *colly.HTMLElement) {
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
}

// extractOtherMetadata extracts description, section, category, and canonical URL.
func extractOtherMetadata(
	data *ArticleData,
	e *colly.HTMLElement,
	selectors configtypes.ArticleSelectors,
	sourceURL string,
) {
	data.Description = extractMetaName(e, "description")
	if data.Description == "" {
		data.Description = data.Intro
	}

	data.Section = extractText(e, selectors.Section)
	if data.Section == "" {
		data.Section = extractMeta(e, "article:section")
	}

	// Extract category
	data.Category = extractText(e, selectors.Category)
	if data.Category == "" {
		data.Category = extractMeta(e, "article:section")
	}
	// Clean category will be applied when converting to domain.Article

	data.CanonicalURL = extractAttr(e, selectors.Canonical, "href")
	if data.CanonicalURL == "" {
		data.CanonicalURL = sourceURL
	}
}

// extractPublishedDate extracts published date with multiple fallback strategies
func extractPublishedDate(e *colly.HTMLElement, selectors configtypes.ArticleSelectors) time.Time {
	// Strategy 1: Try JSON-LD structured data (highest priority)
	if selectors.JSONLD != "" {
		if date := extractDateFromJSONLD(e, selectors.JSONLD); !date.IsZero() {
			return date
		}
	}

	// Strategy 2: Try schema.org datePublished in microdata
	if date := extractDateFromSchemaOrg(e); !date.IsZero() {
		return date
	}

	// Strategy 3: Try published_time selector (datetime attribute)
	if date := tryPublishedTimeSelector(e, selectors); !date.IsZero() {
		return date
	}

	// Strategy 4: Try Open Graph article:published_time
	if date := tryOpenGraphDate(e); !date.IsZero() {
		return date
	}

	// Strategy 5: Try meta name="date" or "publishdate"
	if date := tryMetaNameDate(e); !date.IsZero() {
		return date
	}

	// Strategy 6: Try common HTML date patterns
	return tryTimeElementDate(e)
}

// tryPublishedTimeSelector tries to extract date from published_time selector.
func tryPublishedTimeSelector(e *colly.HTMLElement, selectors configtypes.ArticleSelectors) time.Time {
	// Try datetime attribute first
	publishedTimeStr := extractAttr(e, selectors.PublishedTime, "datetime")
	if publishedTimeStr != "" {
		if date := parseDate(publishedTimeStr); !date.IsZero() {
			return date
		}
	}

	// Try text content
	publishedTimeStr = extractText(e, selectors.PublishedTime)
	if publishedTimeStr != "" {
		if date := parseDate(publishedTimeStr); !date.IsZero() {
			return date
		}
	}

	return time.Time{}
}

// tryOpenGraphDate tries to extract date from Open Graph meta tag.
func tryOpenGraphDate(e *colly.HTMLElement) time.Time {
	publishedTimeStr := extractMeta(e, "article:published_time")
	if publishedTimeStr != "" {
		if date := parseDate(publishedTimeStr); !date.IsZero() {
			return date
		}
	}
	return time.Time{}
}

// tryMetaNameDate tries to extract date from meta name tags.
func tryMetaNameDate(e *colly.HTMLElement) time.Time {
	metaNames := []string{"date", "publishdate", "pubdate"}
	for _, name := range metaNames {
		publishedTimeStr := extractMetaName(e, name)
		if publishedTimeStr != "" {
			if date := parseDate(publishedTimeStr); !date.IsZero() {
				return date
			}
		}
	}
	return time.Time{}
}

// tryTimeElementDate tries to extract date from time element.
func tryTimeElementDate(e *colly.HTMLElement) time.Time {
	publishedTimeStr := extractAttr(e, "time", "datetime")
	if publishedTimeStr != "" {
		if date := parseDate(publishedTimeStr); !date.IsZero() {
			return date
		}
	}
	return time.Time{}
}

// extractDateFromJSONLD extracts published date from JSON-LD structured data
func extractDateFromJSONLD(e *colly.HTMLElement, selector string) time.Time {
	if selector == "" {
		return time.Time{}
	}

	scripts := e.DOM.Find(selector)
	for i := range scripts.Length() {
		script := scripts.Eq(i)
		jsonText := script.Text()
		if jsonText == "" {
			continue
		}

		date := parseJSONLDDate(jsonText)
		if !date.IsZero() {
			return date
		}
	}

	return time.Time{}
}

// parseJSONLDDate parses JSON-LD text and extracts date.
func parseJSONLDDate(jsonText string) time.Time {
	var jsonData any
	if err := json.Unmarshal([]byte(jsonText), &jsonData); err != nil {
		return time.Time{}
	}

	items := normalizeJSONLDItems(jsonData)
	return findDateInJSONLDItems(items)
}

// normalizeJSONLDItems normalizes JSON-LD data to a slice of items.
func normalizeJSONLDItems(jsonData any) []any {
	switch v := jsonData.(type) {
	case []any:
		return v
	case map[string]any:
		return []any{v}
	default:
		return nil
	}
}

// findDateInJSONLDItems searches for date in JSON-LD items.
func findDateInJSONLDItems(items []any) time.Time {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Try nested @graph first
		if date := extractDateFromGraph(obj); !date.IsZero() {
			return date
		}

		// Check for article type and extract date
		if isArticleType(obj) {
			if date := extractDateFromJSONLDObject(obj); !date.IsZero() {
				return date
			}
		}
	}

	return time.Time{}
}

// extractDateFromGraph extracts date from @graph in JSON-LD.
func extractDateFromGraph(obj map[string]any) time.Time {
	// Only process if @type is not present (meaning we should check @graph)
	if _, hasType := obj["@type"].(string); hasType {
		return time.Time{}
	}

	graph, hasGraph := obj["@graph"].([]any)
	if !hasGraph {
		return time.Time{}
	}

	for _, graphItem := range graph {
		graphObj, isGraphObj := graphItem.(map[string]any)
		if !isGraphObj {
			continue
		}
		if date := extractDateFromJSONLDObject(graphObj); !date.IsZero() {
			return date
		}
	}

	return time.Time{}
}

// isArticleType checks if JSON-LD object is an article type.
func isArticleType(obj map[string]any) bool {
	typeVal, hasType := obj["@type"].(string)
	if !hasType {
		return false
	}

	articleTypes := []string{"NewsArticle", "Article", "BlogPosting", "ScholarlyArticle", "Report"}
	for _, articleType := range articleTypes {
		if typeVal == articleType {
			return true
		}
	}

	return false
}

// extractDateFromJSONLDObject extracts date from a JSON-LD object
func extractDateFromJSONLDObject(obj map[string]any) time.Time {
	// Try datePublished first
	if datePublished, ok := obj["datePublished"].(string); ok {
		if date := parseDate(datePublished); !date.IsZero() {
			return date
		}
	}

	// Try publishedDate
	if publishedDate, ok := obj["publishedDate"].(string); ok {
		if date := parseDate(publishedDate); !date.IsZero() {
			return date
		}
	}

	// Try date
	if date, ok := obj["date"].(string); ok {
		if parsedDate := parseDate(date); !parsedDate.IsZero() {
			return parsedDate
		}
	}

	return time.Time{}
}

// extractDateFromSchemaOrg extracts date from schema.org microdata
func extractDateFromSchemaOrg(e *colly.HTMLElement) time.Time {
	// Try itemscope with itemtype="http://schema.org/NewsArticle" or similar
	articleTypes := []string{
		"http://schema.org/NewsArticle",
		"http://schema.org/Article",
		"https://schema.org/NewsArticle",
		"https://schema.org/Article",
	}

	for _, articleType := range articleTypes {
		selector := fmt.Sprintf("[itemtype='%s']", articleType)
		article := e.DOM.Find(selector).First()
		if article.Length() == 0 {
			continue
		}

		if date := extractDateFromSchemaArticle(article); !date.IsZero() {
			return date
		}
	}

	return time.Time{}
}

// extractDateFromSchemaArticle extracts date from a schema.org article element.
func extractDateFromSchemaArticle(article *goquery.Selection) time.Time {
	datePublished := article.Find("[itemprop='datePublished']").First()
	if datePublished.Length() == 0 {
		return time.Time{}
	}

	dateStr := datePublished.AttrOr("content", datePublished.AttrOr("datetime", datePublished.Text()))
	if dateStr == "" {
		return time.Time{}
	}

	date := parseDate(dateStr)
	if date.IsZero() {
		return time.Time{}
	}

	return date
}

// extractArticleID extracts the article ID from various attributes or generates one from URL.
func extractArticleID(e *colly.HTMLElement, selectors configtypes.ArticleSelectors, sourceURL string) string {
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
	return articleID
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
	Keywords      []string
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
