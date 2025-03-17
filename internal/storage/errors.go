package storage

import "errors"

var (
	// ErrInvalidHits indicates hits field is missing or invalid in response
	ErrInvalidHits = errors.New("invalid response format: hits not found")
	// ErrInvalidHitsArray indicates hits array is missing or invalid
	ErrInvalidHitsArray = errors.New("invalid response format: hits array not found")
	// ErrMissingURL indicates the Elasticsearch URL is not configured
	ErrMissingURL = errors.New("elasticsearch URL is required")
	// ErrInvalidScrollID indicates an invalid or missing scroll ID in response
	ErrInvalidScrollID = errors.New("invalid scroll ID")
	// ErrIndexNotFound indicates the requested index does not exist
	ErrIndexNotFound = errors.New("index not found")
	// ErrInvalidIndexHealth indicates the index health is invalid
	ErrInvalidIndexHealth = errors.New("invalid index health format")
	// ErrInvalidDocCount indicates the index document count is invalid
	ErrInvalidDocCount = errors.New("invalid index document count format")
)
