// Package domain provides domain models used across the application.
package domain

// Type represents the type of content being processed.
type Type string

const (
	// TypeArticle represents article content.
	TypeArticle Type = "article"
	// TypePage represents generic page content.
	TypePage Type = "page"
	// TypeVideo represents video content.
	TypeVideo Type = "video"
	// TypeImage represents image content.
	TypeImage Type = "image"
	// TypeHTML represents HTML content.
	TypeHTML Type = "html"
	// TypeJob represents job content.
	TypeJob Type = "job"
)

// These types are defined as any to avoid import cycles.
// They will be used by other packages that need these types.
type (
	// Config represents the configuration interface.
	Config any

	// Storage represents the storage interface.
	Storage any
)
