package models

import "time"

// Content represents a generic web content item
type Content struct {
	ID        string                 `json:"id"`
	URL       string                 `json:"url"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}
