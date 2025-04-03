package models

import "time"

type Article struct {
	// Unique identifier for the article
	ID string `json:"id" mapstructure:"id"`
	// Title of the article
	Title string `json:"title" mapstructure:"title"`
	// Main content of the article
	Body string `json:"body" mapstructure:"body"`
	// Author of the article
	Author string `json:"author" mapstructure:"author"`
	// Byline name if different from author
	BylineName string `json:"byline_name" mapstructure:"byline_name"`
	// Date when the article was published
	PublishedDate time.Time `json:"published_date" mapstructure:"published_date"`
	// Source of the article (e.g., website URL)
	Source string `json:"source" mapstructure:"source"`
	// Tags or categories related to the article
	Tags []string `json:"tags" mapstructure:"tags"`
	// Article introduction or summary
	Intro string `json:"intro" mapstructure:"intro"`
	// Article description (often from meta tags)
	Description string `json:"description" mapstructure:"description"`

	// Open Graph metadata
	// Open Graph title
	OgTitle string `json:"og_title" mapstructure:"og_title"`
	// Open Graph description
	OgDescription string `json:"og_description" mapstructure:"og_description"`
	// Open Graph image URL
	OgImage string `json:"og_image" mapstructure:"og_image"`
	// Open Graph URL
	OgURL string `json:"og_url" mapstructure:"og_url"`

	// Additional metadata
	// Canonical URL if different from source
	CanonicalURL string `json:"canonical_url" mapstructure:"canonical_url"`
	// Article word count
	WordCount int `json:"word_count" mapstructure:"word_count"`
	// Primary category
	Category string `json:"category" mapstructure:"category"`
	// Article section
	Section string `json:"section" mapstructure:"section"`
	// Keywords from meta tags
	Keywords []string `json:"keywords" mapstructure:"keywords"`
	// Record creation timestamp
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	// Record update timestamp
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
