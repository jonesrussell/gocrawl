package events

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/content"
)

// EventType represents the type of event.
type EventType string

const (
	// EventTypeJobStarted is emitted when a job starts.
	EventTypeJobStarted EventType = "job_started"
	// EventTypeJobCompleted is emitted when a job completes successfully.
	EventTypeJobCompleted EventType = "job_completed"
	// EventTypeJobFailed is emitted when a job fails.
	EventTypeJobFailed EventType = "job_failed"
	// EventTypeJobProgress is emitted when job progress is updated.
	EventTypeJobProgress EventType = "job_progress"
)

// Event represents a job-related event.
type Event struct {
	// Type is the type of event.
	Type EventType `json:"type"`
	// Job is the job associated with the event.
	Job *content.Job `json:"job"`
	// Error is the error message if the event is job_failed.
	Error string `json:"error,omitempty"`
	// Progress is the job progress (0-100) if the event is job_progress.
	Progress int `json:"progress,omitempty"`
}

// Handler handles job-related events.
type Handler interface {
	// HandleEvent processes a job event.
	HandleEvent(ctx context.Context, event Event) error
}

// Publisher publishes job-related events.
type Publisher interface {
	// PublishEvent publishes a job event.
	PublishEvent(ctx context.Context, event Event) error
}

// Bus represents an event bus for job-related events.
type Bus interface {
	Publisher
	// Subscribe registers an event handler.
	Subscribe(handler Handler) error
	// Unsubscribe removes an event handler.
	Unsubscribe(handler Handler) error
}
