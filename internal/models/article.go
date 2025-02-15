package models

import "time"

type Article struct {
	ID            string    `json:"id"`             // Unique identifier for the article
	Title         string    `json:"title"`          // Title of the article
	Body          string    `json:"body"`           // Main content of the article
	Author        string    `json:"author"`         // Author of the article
	PublishedDate time.Time `json:"published_date"` // Date when the article was published
	Source        string    `json:"source"`         // Source of the article (e.g., website URL)
	Tags          []string  `json:"tags"`           // Tags or categories related to the article
}
