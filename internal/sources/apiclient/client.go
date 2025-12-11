// Package apiclient provides HTTP client functionality for interacting with the gosources API.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the default base URL for the gosources API.
	DefaultBaseURL = "http://localhost:8050/api/v1/sources"
	// DefaultTimeout is the default timeout for API requests.
	DefaultTimeout = 30 * time.Second
)

// Client is an HTTP client for interacting with the gosources API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option is a function that configures a Client.
type Option func(*Client)

// WithBaseURL sets the base URL for the API client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the timeout for API requests.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new gosources API client.
func NewClient(opts ...Option) *Client {
	client := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// ListSources retrieves all sources from the API.
func (c *Client) ListSources(ctx context.Context) ([]APISource, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response ListSourcesResponse
	if doErr := c.doRequest(req, &response); doErr != nil {
		return nil, fmt.Errorf("failed to list sources: %w", doErr)
	}

	return response.Sources, nil
}

// GetSource retrieves a specific source by ID.
func (c *Client) GetSource(ctx context.Context, id string) (*APISource, error) {
	sourceURL, err := url.JoinPath(c.baseURL, id)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var source APISource
	if doErr := c.doRequest(req, &source); doErr != nil {
		return nil, fmt.Errorf("failed to get source: %w", doErr)
	}

	return &source, nil
}

// CreateSource creates a new source via the API.
func (c *Client) CreateSource(ctx context.Context, source *APISource) (*APISource, error) {
	body, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal source: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var created APISource
	if doErr := c.doRequest(req, &created); doErr != nil {
		return nil, fmt.Errorf("failed to create source: %w", doErr)
	}

	return &created, nil
}

// UpdateSource updates an existing source via the API.
func (c *Client) UpdateSource(ctx context.Context, id string, source *APISource) (*APISource, error) {
	sourceURL, err := url.JoinPath(c.baseURL, id)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}

	body, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal source: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, sourceURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var updated APISource
	if doErr := c.doRequest(req, &updated); doErr != nil {
		return nil, fmt.Errorf("failed to update source: %w", doErr)
	}

	return &updated, nil
}

// DeleteSource deletes a source via the API.
func (c *Client) DeleteSource(ctx context.Context, id string) error {
	sourceURL, err := url.JoinPath(c.baseURL, id)
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, sourceURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if doErr := c.doRequest(req, nil); doErr != nil {
		return fmt.Errorf("failed to delete source: %w", doErr)
	}

	return nil
}

// doRequest executes an HTTP request and decodes the response.
func (c *Client) doRequest(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Provide more helpful error message for connection issues
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if urlErr.Op == "dial" || urlErr.Err != nil {
				return fmt.Errorf("failed to connect to sources API at %s: %w. "+
					"Ensure the gosources service is running and accessible", c.baseURL, err)
			}
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("failed to read response body: %w", readErr)
	}

	// Check for error status codes
	const minErrorStatusCode = 400
	if resp.StatusCode >= minErrorStatusCode {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error != "" {
			return fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Error, errResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// For DELETE requests with 204 No Content, don't try to decode
	if resp.StatusCode == http.StatusNoContent || result == nil {
		return nil
	}

	// Decode the response
	if unmarshalErr := json.Unmarshal(body, result); unmarshalErr != nil {
		return fmt.Errorf("failed to decode response: %w", unmarshalErr)
	}

	return nil
}
