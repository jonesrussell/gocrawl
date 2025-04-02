// Package types provides common type definitions used across the application.
package types

import (
	"time"
)

// Storage defines the interface for data storage operations.
type Storage interface {
	// Add your storage methods here
}

// Config defines the interface for configuration operations.
type Config interface {
	// Add your config methods here
}

// Job represents a crawling job.
type Job struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Item represents a crawled item from a job.
type Item struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
