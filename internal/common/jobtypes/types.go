package jobtypes

import "time"

// JobState represents the state of a job.
type JobState string

const (
	// JobStateRunning indicates the job is currently running.
	JobStateRunning JobState = "running"
	// JobStateCompleted indicates the job has completed successfully.
	JobStateCompleted JobState = "completed"
	// JobStateFailed indicates the job has failed.
	JobStateFailed JobState = "failed"
	// JobStatePending indicates the job is waiting to start.
	JobStatePending JobState = "pending"
	// JobStateStopped indicates the job has been manually stopped.
	JobStateStopped JobState = "stopped"
)

// JobStatus represents the current status of a job.
type JobStatus struct {
	// State is the current state of the job.
	State JobState `json:"state"`
	// StartTime is when the job started.
	StartTime time.Time `json:"start_time"`
	// EndTime is when the job completed or failed.
	EndTime time.Time `json:"end_time,omitempty"`
	// Error is the error message if the job failed.
	Error string `json:"error,omitempty"`
	// Progress is the percentage of completion (0-100).
	Progress int `json:"progress"`
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
