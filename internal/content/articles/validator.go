// Package articles provides functionality for processing and managing article content.
package articles

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ValidationResult represents the result of article validation
type ValidationResult struct {
	IsValid bool
	Reason  string
}

// ArticleValidator validates articles before indexing
type ArticleValidator struct {
	logger logger.Interface
	stats  *ValidationStats
}

// ValidationStats tracks validation statistics
type ValidationStats struct {
	TotalProcessed      int64
	ValidArticles       int64
	InvalidArticles     int64
	CategoryPageSkips   int64
	InvalidDateSkips    int64
	ContentQualitySkips int64
	TitleQualitySkips   int64
	WordCountSkips      int64
}

// NewArticleValidator creates a new article validator
func NewArticleValidator(log logger.Interface) *ArticleValidator {
	return &ArticleValidator{
		logger: log,
		stats:  &ValidationStats{},
	}
}

// GetStats returns validation statistics
func (v *ArticleValidator) GetStats() ValidationStats {
	return *v.stats
}

// ResetStats resets validation statistics
func (v *ArticleValidator) ResetStats() {
	v.stats = &ValidationStats{}
}

// Skip patterns for category/listing pages
var skipPatterns = []string{
	"/category/",
	"/tag/",
	"/page/",
	"/author/",
	"/archive/",
	"/feed/",
	"/rss/",
	"/search/",
	"?page=",
	"/page/",
}

// Generic titles that indicate category/listing pages
var genericTitles = []string{
	"latest headlines",
	"latest news",
	"news archive",
	"headlines",
	"news",
	"articles",
	"all articles",
	"category",
	"tag",
	"archive",
}

// ValidateArticle validates an article before indexing
func (v *ArticleValidator) ValidateArticle(article *domain.Article) ValidationResult {
	v.stats.TotalProcessed++

	if article == nil {
		v.stats.InvalidArticles++
		return ValidationResult{
			IsValid: false,
			Reason:  "article is nil",
		}
	}

	// Check 1: Category page detection
	if result := v.isCategoryPage(article); !result.IsValid {
		v.stats.InvalidArticles++
		v.stats.CategoryPageSkips++
		v.logger.Debug("Article validation failed: category page",
			"url", article.Source,
			"reason", result.Reason)
		return result
	}

	// Check 2: Invalid published date
	if result := v.validatePublishedDate(article); !result.IsValid {
		v.stats.InvalidArticles++
		v.stats.InvalidDateSkips++
		v.logger.Debug("Article validation failed: invalid date",
			"url", article.Source,
			"reason", result.Reason)
		return result
	}

	// Check 3: Content quality
	if result := v.validateContent(article); !result.IsValid {
		v.stats.InvalidArticles++
		v.stats.ContentQualitySkips++
		v.logger.Debug("Article validation failed: content quality",
			"url", article.Source,
			"reason", result.Reason)
		return result
	}

	// Check 4: Title quality
	if result := v.validateTitle(article); !result.IsValid {
		v.stats.InvalidArticles++
		v.stats.TitleQualitySkips++
		v.logger.Debug("Article validation failed: title quality",
			"url", article.Source,
			"reason", result.Reason)
		return result
	}

	// Check 5: Word count
	if result := v.validateWordCount(article); !result.IsValid {
		v.stats.InvalidArticles++
		v.stats.WordCountSkips++
		v.logger.Debug("Article validation failed: word count",
			"url", article.Source,
			"reason", result.Reason)
		return result
	}

	v.stats.ValidArticles++
	return ValidationResult{IsValid: true}
}

// isCategoryPage checks if the URL or content indicates a category/listing page
func (v *ArticleValidator) isCategoryPage(article *domain.Article) ValidationResult {
	sourceURL := article.Source
	if sourceURL == "" {
		sourceURL = article.CanonicalURL
	}

	// Check URL patterns
	parsedURL, err := url.Parse(sourceURL)
	if err == nil {
		path := strings.ToLower(parsedURL.Path)
		query := strings.ToLower(parsedURL.RawQuery)

		// Check for skip patterns in path
		for _, pattern := range skipPatterns {
			if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
				return ValidationResult{
					IsValid: false,
					Reason:  fmt.Sprintf("URL matches skip pattern: %s", pattern),
				}
			}
		}

		// Check if URL ends with / (often indicates category pages)
		if path != "/" && strings.HasSuffix(path, "/") {
			// Allow root path but skip other trailing slashes
			return ValidationResult{
				IsValid: false,
				Reason:  "URL ends with trailing slash (likely category page)",
			}
		}
	}

	// Check for generic titles
	titleLower := strings.ToLower(strings.TrimSpace(article.Title))
	for _, generic := range genericTitles {
		if titleLower == generic ||
			strings.HasPrefix(titleLower, generic+" |") ||
			strings.HasSuffix(titleLower, "| "+generic) {
			return ValidationResult{
				IsValid: false,
				Reason:  fmt.Sprintf("Generic title detected: %s", article.Title),
			}
		}
	}

	// Check for concatenated content (multiple article snippets)
	// Check body, intro, and description as they can all contain concatenated content
	if v.hasConcatenatedContent(article.Body) {
		return ValidationResult{
			IsValid: false,
			Reason:  "Content appears to be concatenated snippets from multiple articles (body)",
		}
	}
	if article.Intro != "" && v.hasConcatenatedContent(article.Intro) {
		return ValidationResult{
			IsValid: false,
			Reason:  "Content appears to be concatenated snippets from multiple articles (intro)",
		}
	}
	if article.Description != "" && v.hasConcatenatedContent(article.Description) {
		return ValidationResult{
			IsValid: false,
			Reason:  "Content appears to be concatenated snippets from multiple articles (description)",
		}
	}

	return ValidationResult{IsValid: true}
}

const (
	minBodyLengthForConcatenation = 200
	minSeparatorCount             = 3
	minHeadlineLikeCount          = 5
	minContentLength              = 100
	maxContentLength              = 100000
	minWordCount                  = 50
)

// hasConcatenatedContent checks if body contains multiple article snippets
func (v *ArticleValidator) hasConcatenatedContent(body string) bool {
	if len(body) < minBodyLengthForConcatenation {
		return false // Too short to be concatenated
	}

	// Look for patterns that suggest multiple articles:
	// - Multiple "Read more" or "Continue reading" patterns
	// - Multiple article-like structures
	// - Excessive repetition of similar patterns

	// Count occurrences of common article separators
	separators := []string{
		"read more",
		"continue reading",
		"full story",
		"view article",
	}

	bodyLower := strings.ToLower(body)
	separatorCount := 0
	for _, sep := range separators {
		separatorCount += strings.Count(bodyLower, sep)
	}

	// If we find 3+ separators, likely multiple articles
	if separatorCount >= minSeparatorCount {
		return true
	}

	// Check for multiple headline-like patterns (lines that are very short and end with punctuation)
	lines := strings.Split(body, "\n")
	headlineLikeCount := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Headline-like: short line (20-100 chars) ending with punctuation
		if len(line) >= 20 && len(line) <= 100 {
			if strings.HasSuffix(line, ".") || strings.HasSuffix(line, "?") || strings.HasSuffix(line, "!") {
				// Check if it looks like a headline (capitalized, no lowercase words in middle)
				if len(strings.Fields(line)) >= 3 && len(strings.Fields(line)) <= 15 {
					headlineLikeCount++
				}
			}
		}
	}

	// If we find 5+ headline-like patterns, likely multiple articles
	if headlineLikeCount >= minHeadlineLikeCount {
		return true
	}

	return false
}

// validatePublishedDate validates the published date
func (v *ArticleValidator) validatePublishedDate(article *domain.Article) ValidationResult {
	// Check for zero-value date
	if article.PublishedDate.IsZero() {
		return ValidationResult{
			IsValid: false,
			Reason:  "Published date is zero-value (0001-01-01T00:00:00Z)",
		}
	}

	// Validate date is reasonable (between 2000 and current date + 1 day)
	now := time.Now()
	minDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := now.AddDate(0, 0, 1) // Allow 1 day in future for timezone issues

	if article.PublishedDate.Before(minDate) {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Published date is before 2000: %s", article.PublishedDate.Format(time.RFC3339)),
		}
	}

	if article.PublishedDate.After(maxDate) {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Published date is too far in future: %s", article.PublishedDate.Format(time.RFC3339)),
		}
	}

	return ValidationResult{IsValid: true}
}

// validateContent validates content quality
func (v *ArticleValidator) validateContent(article *domain.Article) ValidationResult {
	body := strings.TrimSpace(article.Body)

	// Check minimum length
	if len(body) < minContentLength {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Content too short: %d characters (minimum %d)", len(body), minContentLength),
		}
	}

	// Check maximum length (likely concatenated content)
	if len(body) > maxContentLength {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Content too long: %d characters (maximum %d)", len(body), maxContentLength),
		}
	}

	return ValidationResult{IsValid: true}
}

// validateTitle validates title quality
func (v *ArticleValidator) validateTitle(article *domain.Article) ValidationResult {
	title := strings.TrimSpace(article.Title)

	if title == "" {
		return ValidationResult{
			IsValid: false,
			Reason:  "Title is empty",
		}
	}

	// Check for generic titles
	titleLower := strings.ToLower(title)
	for _, generic := range genericTitles {
		if titleLower == generic {
			return ValidationResult{
				IsValid: false,
				Reason:  fmt.Sprintf("Generic title: %s", title),
			}
		}
	}

	return ValidationResult{IsValid: true}
}

// validateWordCount validates word count
func (v *ArticleValidator) validateWordCount(article *domain.Article) ValidationResult {
	// Word count should be calculated, but if it's 0, we'll calculate it
	if article.WordCount == 0 {
		wordCount := CalculateWordCount(article.Body)
		if wordCount < minWordCount {
			return ValidationResult{
				IsValid: false,
				Reason:  fmt.Sprintf("Word count too low: %d words (minimum %d)", wordCount, minWordCount),
			}
		}
	} else if article.WordCount < minWordCount {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Word count too low: %d words (minimum %d)", article.WordCount, minWordCount),
		}
	}

	return ValidationResult{IsValid: true}
}

// CalculateWordCount calculates word count from text, stripping HTML tags
func CalculateWordCount(text string) int {
	if text == "" {
		return 0
	}

	// Remove HTML tags using regex
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned := htmlTagRegex.ReplaceAllString(text, " ")

	// Remove extra whitespace
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// Split into words
	words := strings.Fields(cleaned)
	return len(words)
}

// CleanCategory cleans and normalizes category field
func CleanCategory(category string) []string {
	if category == "" {
		return []string{}
	}

	// Remove excessive whitespace
	category = regexp.MustCompile(`\s+`).ReplaceAllString(category, " ")
	category = strings.TrimSpace(category)

	// Split by common separators
	separators := []string{",", "|", "/", "\\", "•", "·"}
	categories := []string{category}

	for _, sep := range separators {
		var newCategories []string
		for _, cat := range categories {
			parts := strings.Split(cat, sep)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					newCategories = append(newCategories, part)
				}
			}
		}
		categories = newCategories
	}

	// Remove duplicates
	seen := make(map[string]bool)
	result := []string{}
	for _, cat := range categories {
		catLower := strings.ToLower(strings.TrimSpace(cat))
		if !seen[catLower] && catLower != "" {
			seen[catLower] = true
			result = append(result, strings.TrimSpace(cat))
		}
	}

	// If we ended up with just repeated values, return empty
	if len(result) == 1 && strings.Count(category, result[0]) > 3 {
		// Likely just repeated value like "Canada    Canada    Canada"
		return []string{}
	}

	return result
}
