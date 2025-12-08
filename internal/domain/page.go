// Package domain provides domain models used across the application.
package domain

import (
	"time"

	"github.com/gocolly/colly/v2"
)

// Page represents a web page.
type Page struct {
	// Unique identifier for the page
	ID string `json:"id" mapstructure:"id"`
	// URL of the page
	URL string `json:"url" mapstructure:"url"`
	// Title of the page
	Title string `json:"title" mapstructure:"title"`
	// Main content of the page
	Content string `json:"content" mapstructure:"content"`
	// Description of the page (often from meta tags)
	Description string `json:"description" mapstructure:"description"`
	// Keywords from meta tags
	Keywords []string `json:"keywords" mapstructure:"keywords"`
	// Raw HTML content
	HTML *colly.HTMLElement `json:"-" mapstructure:"-"`
	// Open Graph metadata
	OgTitle       string `json:"og_title" mapstructure:"og_title"`
	OgDescription string `json:"og_description" mapstructure:"og_description"`
	OgImage       string `json:"og_image" mapstructure:"og_image"`
	OgURL         string `json:"og_url" mapstructure:"og_url"`
	// Canonical URL if different from source
	CanonicalURL string `json:"canonical_url" mapstructure:"canonical_url"`
	// Record creation timestamp
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	// Record update timestamp
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
