// Package types provides common types used across the application.
package types

// APIError represents an error response from the API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}
