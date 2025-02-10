package cmd

import "time"

// Shared command variables
var (
	baseURL   string
	maxDepth  int
	rateLimit time.Duration
	indexName string // Shared between search and crawl commands
	batchSize int
	querySize int
)
