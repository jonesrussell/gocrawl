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
)
