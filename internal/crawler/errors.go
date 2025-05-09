// Package crawler provides the core crawling functionality for GoCrawl.
package crawler

import (
	"errors"
	"fmt"
)

// Error types for the crawler package.
var (
	// ErrSourceNotFound is returned when the requested source is not found.
	ErrSourceNotFound = errors.New("source not found")

	// ErrIndexNotFound is returned when the requested index is not found.
	ErrIndexNotFound = errors.New("index not found")

	// ErrInvalidConfig is returned when the crawler configuration is invalid.
	ErrInvalidConfig = errors.New("invalid crawler configuration")

	// ErrRateLimitExceeded is returned when the rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrMaxDepthReached is returned when the maximum depth is reached.
	ErrMaxDepthReached = errors.New("maximum depth reached")

	// ErrForbiddenDomain is returned when the domain is not allowed.
	ErrForbiddenDomain = errors.New("forbidden domain")

	// ErrInvalidURL is returned when the URL is invalid.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrContentProcessingFailed is returned when content processing fails.
	ErrContentProcessingFailed = errors.New("content processing failed")

	// ErrArticleProcessingFailed is returned when article processing fails.
	ErrArticleProcessingFailed = errors.New("article processing failed")

	// ErrAlreadyVisited is returned when a URL has already been visited
	ErrAlreadyVisited = errors.New("URL already visited")

	// ErrMaxDepth is returned when the maximum crawl depth has been reached
	ErrMaxDepth = errors.New("maximum crawl depth reached")

	// ErrMissingURL is returned when a URL is missing or invalid
	ErrMissingURL = errors.New("missing or invalid URL")
)

// WrapperError wraps an error with additional context.
type WrapperError struct {
	Err     error
	Context string
}

// Error returns the error message.
func (e *WrapperError) Error() string {
	return fmt.Sprintf("%s: %v", e.Context, e.Err)
}

// Unwrap returns the underlying error.
func (e *WrapperError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with additional context.
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return &WrapperError{
		Err:     err,
		Context: context,
	}
}

// IsSourceNotFoundError returns true if the error is a source not found error.
func IsSourceNotFoundError(err error) bool {
	return errors.Is(err, ErrSourceNotFound)
}

// IsIndexNotFoundError returns true if the error is an index not found error.
func IsIndexNotFoundError(err error) bool {
	return errors.Is(err, ErrIndexNotFound)
}

// IsInvalidConfigError returns true if the error is an invalid config error.
func IsInvalidConfigError(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

// IsRateLimitExceededError returns true if the error is a rate limit exceeded error.
func IsRateLimitExceededError(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

// IsMaxDepthReachedError returns true if the error is a max depth reached error.
func IsMaxDepthReachedError(err error) bool {
	return errors.Is(err, ErrMaxDepthReached)
}

// IsForbiddenDomainError returns true if the error is a forbidden domain error.
func IsForbiddenDomainError(err error) bool {
	return errors.Is(err, ErrForbiddenDomain)
}

// IsInvalidURLError returns true if the error is an invalid URL error.
func IsInvalidURLError(err error) bool {
	return errors.Is(err, ErrInvalidURL)
}

// IsContentProcessingFailedError returns true if the error is a content processing failed error.
func IsContentProcessingFailedError(err error) bool {
	return errors.Is(err, ErrContentProcessingFailed)
}

// IsArticleProcessingFailedError returns true if the error is an article processing failed error.
func IsArticleProcessingFailedError(err error) bool {
	return errors.Is(err, ErrArticleProcessingFailed)
}

// NewWrapperError creates a new error wrapper.
func NewWrapperError(err error, context string) error {
	return &WrapperError{
		Err:     err,
		Context: context,
	}
}
