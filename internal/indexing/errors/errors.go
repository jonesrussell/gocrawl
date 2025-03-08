// Package errors provides custom error types for indexing operations.
package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrIndexNotFound indicates that the requested index does not exist.
	ErrIndexNotFound = errors.New("index not found")

	// ErrIndexAlreadyExists indicates that an index with the given name already exists.
	ErrIndexAlreadyExists = errors.New("index already exists")

	// ErrDocumentNotFound indicates that the requested document does not exist.
	ErrDocumentNotFound = errors.New("document not found")

	// ErrInvalidMapping indicates that the provided mapping is invalid.
	ErrInvalidMapping = errors.New("invalid mapping")

	// ErrInvalidQuery indicates that the provided query is invalid.
	ErrInvalidQuery = errors.New("invalid query")
)

// IndexError represents an error that occurred during index operations.
type IndexError struct {
	Index string
	Op    string
	Err   error
}

// Error implements the error interface.
func (e *IndexError) Error() string {
	return fmt.Sprintf("index %q: operation %q failed: %v", e.Index, e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *IndexError) Unwrap() error {
	return e.Err
}

// DocumentError represents an error that occurred during document operations.
type DocumentError struct {
	Index string
	ID    string
	Op    string
	Err   error
}

// Error implements the error interface.
func (e *DocumentError) Error() string {
	return fmt.Sprintf("document %q in index %q: operation %q failed: %v", e.ID, e.Index, e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *DocumentError) Unwrap() error {
	return e.Err
}

// SearchError represents an error that occurred during search operations.
type SearchError struct {
	Index string
	Query interface{}
	Op    string
	Err   error
}

// Error implements the error interface.
func (e *SearchError) Error() string {
	return fmt.Sprintf("search in index %q: operation %q failed: %v", e.Index, e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *SearchError) Unwrap() error {
	return e.Err
}

// NewIndexError creates a new IndexError.
func NewIndexError(index, op string, err error) error {
	return &IndexError{
		Index: index,
		Op:    op,
		Err:   err,
	}
}

// NewDocumentError creates a new DocumentError.
func NewDocumentError(index, id, op string, err error) error {
	return &DocumentError{
		Index: index,
		ID:    id,
		Op:    op,
		Err:   err,
	}
}

// NewSearchError creates a new SearchError.
func NewSearchError(index string, query interface{}, op string, err error) error {
	return &SearchError{
		Index: index,
		Query: query,
		Op:    op,
		Err:   err,
	}
}
