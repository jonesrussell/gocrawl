package models

import "time"

type Article struct {
	ID            string    `json:"id" mapstructure:"id"`                         // Unique identifier for the article
	Title         string    `json:"title" mapstructure:"title"`                   // Title of the article
	Body          string    `json:"body" mapstructure:"body"`                     // Main content of the article
	Author        string    `json:"author" mapstructure:"author"`                 // Author of the article
	BylineName    string    `json:"byline_name" mapstructure:"byline_name"`       // Byline name if different from author
	PublishedDate time.Time `json:"published_date" mapstructure:"published_date"` // Date when the article was published
	Source        string    `json:"source" mapstructure:"source"`                 // Source of the article (e.g., website URL)
	Tags          []string  `json:"tags" mapstructure:"tags"`                     // Tags or categories related to the article
	Intro         string    `json:"intro" mapstructure:"intro"`                   // Article introduction or summary
	Description   string    `json:"description" mapstructure:"description"`       // Article description (often from meta tags)

	// Open Graph metadata
	OgTitle       string `json:"og_title" mapstructure:"og_title"`             // Open Graph title
	OgDescription string `json:"og_description" mapstructure:"og_description"` // Open Graph description
	OgImage       string `json:"og_image" mapstructure:"og_image"`             // Open Graph image URL
	OgURL         string `json:"og_url" mapstructure:"og_url"`                 // Open Graph URL

	// Additional metadata
	CanonicalURL string    `json:"canonical_url" mapstructure:"canonical_url"` // Canonical URL if different from source
	WordCount    int       `json:"word_count" mapstructure:"word_count"`       // Article word count
	Category     string    `json:"category" mapstructure:"category"`           // Primary category
	Section      string    `json:"section" mapstructure:"section"`             // Article section
	Keywords     []string  `json:"keywords" mapstructure:"keywords"`           // Keywords from meta tags
	CreatedAt    time.Time `json:"created_at" mapstructure:"created_at"`       // Record creation timestamp
	UpdatedAt    time.Time `json:"updated_at" mapstructure:"updated_at"`       // Record update timestamp
}
