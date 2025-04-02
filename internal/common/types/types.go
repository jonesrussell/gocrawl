package types

import (
	"time"
)

// Logger is an interface for structured logging capabilities.
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Printf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Sync() error
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
