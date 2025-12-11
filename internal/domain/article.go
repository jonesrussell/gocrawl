// Package domain provides domain models used across the application.
package domain

import (
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Article represents an article.
type Article struct {
	// Unique identifier for the article
	ID string `json:"id" mapstructure:"id"`
	// Title of the article
	Title string `json:"title" mapstructure:"title"`
	// Main content of the article
	Body string `json:"body" mapstructure:"body"`
	// Author of the article
	Author string `json:"author,omitempty" mapstructure:"author"`
	// Byline name if different from author
	BylineName string `json:"byline_name,omitempty" mapstructure:"byline_name"`
	// Date when the article was published
	PublishedDate time.Time `json:"published_date" mapstructure:"published_date"`
	// Source of the article (e.g., website URL)
	Source string `json:"source" mapstructure:"source"`
	// Tags or categories related to the article
	Tags []string `json:"tags,omitempty" mapstructure:"tags"`
	// Article introduction or summary
	Intro string `json:"intro,omitempty" mapstructure:"intro"`
	// Article description (often from meta tags)
	Description string `json:"description,omitempty" mapstructure:"description"`
	// Raw HTML content
	HTML *colly.HTMLElement `json:"-" mapstructure:"-"`

	// Open Graph metadata
	// Open Graph title (may differ from Title for social sharing)
	OgTitle string `json:"og_title,omitempty" mapstructure:"og_title"`
	// Open Graph description (may differ from Description/Intro for social sharing)
	OgDescription string `json:"og_description,omitempty" mapstructure:"og_description"`
	// Open Graph image URL
	OgImage string `json:"og_image,omitempty" mapstructure:"og_image"`
	// Open Graph URL (may differ from CanonicalURL)
	OgURL string `json:"og_url,omitempty" mapstructure:"og_url"`

	// Additional metadata
	// Canonical URL if different from source
	CanonicalURL string `json:"canonical_url,omitempty" mapstructure:"canonical_url"`
	// Article word count
	WordCount int `json:"word_count" mapstructure:"word_count"`
	// Primary category
	Category string `json:"category,omitempty" mapstructure:"category"`
	// Article section
	Section string `json:"section,omitempty" mapstructure:"section"`
	// Keywords from meta tags
	Keywords []string `json:"keywords,omitempty" mapstructure:"keywords"`
	// Record creation timestamp
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	// Record update timestamp
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}

// KeywordsString returns keywords as a comma-separated string for Drupal compatibility.
// Drupal expects field_keywords as a string (long text) field.
func (a *Article) KeywordsString() string {
	if len(a.Keywords) == 0 {
		return ""
	}
	return strings.Join(a.Keywords, ", ")
}

// TagsString returns tags as a comma-separated string.
func (a *Article) TagsString() string {
	if len(a.Tags) == 0 {
		return ""
	}
	return strings.Join(a.Tags, ", ")
}

// GetOGTitle returns the Open Graph title, falling back to Title if OG title is not set.
func (a *Article) GetOGTitle() string {
	if a.OgTitle != "" {
		return a.OgTitle
	}
	return a.Title
}

// GetOGDescription returns the Open Graph description, falling back to
// Description or Intro if OG description is not set.
func (a *Article) GetOGDescription() string {
	if a.OgDescription != "" {
		return a.OgDescription
	}
	if a.Description != "" {
		return a.Description
	}
	return a.Intro
}

// GetOGURL returns the Open Graph URL, falling back to CanonicalURL or Source if OG URL is not set.
func (a *Article) GetOGURL() string {
	if a.OgURL != "" {
		return a.OgURL
	}
	if a.CanonicalURL != "" {
		return a.CanonicalURL
	}
	return a.Source
}

// PrepareForIndexing cleans and prepares the article for Elasticsearch indexing.
// This ensures best practices: removes empty strings, normalizes arrays, and prevents duplication.
func (a *Article) PrepareForIndexing() {
	a.cleanEmptyStrings()
	a.removeDuplicateOGFields()
	a.cleanOGFields()
	a.normalizeArrays()
}

// cleanEmptyStrings converts empty strings to empty (will be omitted by omitempty).
func (a *Article) cleanEmptyStrings() {
	a.Author = cleanString(a.Author)
	a.BylineName = cleanString(a.BylineName)
	a.Intro = cleanString(a.Intro)
	a.Description = cleanString(a.Description)
	a.OgImage = cleanString(a.OgImage)
	a.CanonicalURL = cleanString(a.CanonicalURL)
}

// cleanString returns empty string if trimmed value is empty.
func cleanString(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return s
}

// removeDuplicateOGFields removes OG fields that duplicate canonical fields.
func (a *Article) removeDuplicateOGFields() {
	if strings.TrimSpace(a.OgTitle) == strings.TrimSpace(a.Title) {
		a.OgTitle = ""
	}
	if strings.TrimSpace(a.OgDescription) == strings.TrimSpace(a.Description) ||
		strings.TrimSpace(a.OgDescription) == strings.TrimSpace(a.Intro) {
		a.OgDescription = ""
	}
	if strings.TrimSpace(a.OgURL) == strings.TrimSpace(a.CanonicalURL) ||
		strings.TrimSpace(a.OgURL) == strings.TrimSpace(a.Source) {
		a.OgURL = ""
	}
}

// cleanOGFields cleans OG and other metadata fields.
func (a *Article) cleanOGFields() {
	a.OgTitle = cleanString(a.OgTitle)
	a.OgDescription = cleanString(a.OgDescription)
	a.OgURL = cleanString(a.OgURL)
	a.Category = cleanString(a.Category)
	a.Section = cleanString(a.Section)
}

// normalizeArrays ensures empty arrays are nil and deduplicates.
func (a *Article) normalizeArrays() {
	a.Tags = normalizeStringArray(a.Tags)
	a.Keywords = normalizeStringArray(a.Keywords)
}

// normalizeStringArray removes empty items, deduplicates, and returns nil if empty.
func normalizeStringArray(arr []string) []string {
	if len(arr) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	cleaned := make([]string, 0, len(arr))
	for _, item := range arr {
		item = strings.TrimSpace(item)
		if item != "" && !seen[item] {
			seen[item] = true
			cleaned = append(cleaned, item)
		}
	}

	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}
