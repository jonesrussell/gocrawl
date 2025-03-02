package jobs

type Job struct {
	SourceName string
	Status     string // e.g., "pending", "running", "completed", "failed"
	Error      error
}
