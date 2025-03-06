package models

import "time"

type Article struct {
	ID            string    `json:"id"`             // Unique identifier for the article
	Title         string    `json:"title"`          // Title of the article
	Body          string    `json:"body"`           // Main content of the article
	Author        string    `json:"author"`         // Author of the article
	BylineName    string    `json:"byline_name"`    // Byline name if different from author
	PublishedDate time.Time `json:"published_date"` // Date when the article was published
	Source        string    `json:"source"`         // Source of the article (e.g., website URL)
	Tags          []string  `json:"tags"`           // Tags or categories related to the article
	Intro         string    `json:"intro"`          // Article introduction or summary
	Description   string    `json:"description"`    // Article description (often from meta tags)

	// Open Graph metadata
	OgTitle       string `json:"og_title"`       // Open Graph title
	OgDescription string `json:"og_description"` // Open Graph description
	OgImage       string `json:"og_image"`       // Open Graph image URL
	OGURL         string `json:"og_url"`         // Open Graph URL

	// Additional metadata
	CanonicalUrl string    `json:"canonical_url"` // Canonical URL if different from source
	WordCount    int       `json:"word_count"`    // Article word count
	Category     string    `json:"category"`      // Primary category
	Section      string    `json:"section"`       // Article section
	Keywords     []string  `json:"keywords"`      // Keywords from meta tags
	CreatedAt    time.Time `json:"created_at"`    // Record creation timestamp
	UpdatedAt    time.Time `json:"updated_at"`    // Record update timestamp
}
