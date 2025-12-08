// Package helpers provides testing utilities for integration tests.
package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertIndexExists checks that an index exists in Elasticsearch.
func AssertIndexExists(t require.TestingT, storage types.Interface, ctx context.Context, index string) {
	exists, err := storage.IndexExists(ctx, index)
	require.NoError(t, err, "failed to check if index exists")
	assert.True(t, exists, "index %s should exist", index)
}

// AssertIndexNotExists checks that an index does not exist in Elasticsearch.
func AssertIndexNotExists(t require.TestingT, storage types.Interface, ctx context.Context, index string) {
	exists, err := storage.IndexExists(ctx, index)
	require.NoError(t, err, "failed to check if index exists")
	assert.False(t, exists, "index %s should not exist", index)
}

const (
	// DefaultHealthCheckInterval is the default interval for health checks.
	DefaultHealthCheckInterval = 100 * time.Millisecond
)

// AssertDocumentIndexed checks that a document is indexed in Elasticsearch.
func AssertDocumentIndexed(t require.TestingT, storage types.Interface, ctx context.Context, index, id string) {
	var doc map[string]any
	err := storage.GetDocument(ctx, index, id, &doc)
	require.NoError(t, err, "failed to get document %s from index %s", id, index)
	assert.NotNil(t, doc, "document %s should exist in index %s", id, index)
}

// AssertDocumentNotIndexed checks that a document is not indexed in Elasticsearch.
func AssertDocumentNotIndexed(t require.TestingT, storage types.Interface, ctx context.Context, index, id string) {
	var doc map[string]any
	err := storage.GetDocument(ctx, index, id, &doc)
	// Document should not exist, so error is expected
	assert.Error(t, err, "document %s should not exist in index %s", id, index)
}

// AssertDocumentCount checks that an index has the expected number of documents.
func AssertDocumentCount(
	t require.TestingT,
	storage types.Interface,
	ctx context.Context,
	index string,
	expectedCount int64,
) {
	count, err := storage.GetIndexDocCount(ctx, index)
	require.NoError(t, err, "failed to get document count for index %s", index)
	assert.Equal(t, expectedCount, count, "index %s should have %d documents, got %d", index, expectedCount, count)
}

// WaitForIndexReady waits for an index to be ready (green status).
func WaitForIndexReady(
	t require.TestingT,
	storage types.Interface,
	ctx context.Context,
	index string,
	timeout time.Duration,
) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		health, err := storage.GetIndexHealth(ctx, index)
		if err == nil && health == "green" {
			return
		}
		time.Sleep(DefaultHealthCheckInterval)
	}
	require.Fail(t, fmt.Sprintf("index %s did not become ready within %v", index, timeout))
}

// AssertSearchResults checks that a search returns the expected number of results.
func AssertSearchResults(
	t require.TestingT,
	storage types.Interface,
	ctx context.Context,
	index string,
	query map[string]any,
	minResults int,
) {
	results, err := storage.Search(ctx, index, query)
	require.NoError(t, err, "failed to search index %s", index)
	assert.GreaterOrEqual(t, len(results), minResults, "search should return at least %d results", minResults)
}

// RetryAssertion retries an assertion function until it succeeds or timeout is reached.
func RetryAssertion(t require.TestingT, timeout, interval time.Duration, fn func() error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = fn()
		if lastErr == nil {
			return
		}
		time.Sleep(interval)
	}
	require.Fail(t, fmt.Sprintf("assertion failed after %v: %v", timeout, lastErr))
}

// WaitForDocumentIndexed polls until document is indexed or times out.
func WaitForDocumentIndexed(
	t require.TestingT,
	storage types.Interface,
	ctx context.Context,
	indexName, docID string,
	timeout time.Duration,
) {
	require.Eventually(t, func() bool {
		var doc map[string]any
		err := storage.GetDocument(ctx, indexName, docID, &doc)
		return err == nil
	}, timeout, DefaultHealthCheckInterval,
		"document %q not indexed in index %q within %v", docID, indexName, timeout)
}

// WaitForDocumentCount polls until document count matches or times out.
func WaitForDocumentCount(
	t require.TestingT,
	storage types.Interface,
	ctx context.Context,
	indexName string,
	expectedCount int,
	timeout time.Duration,
) {
	require.Eventually(t, func() bool {
		count, err := storage.GetIndexDocCount(ctx, indexName)
		if err != nil {
			return false
		}
		return count == int64(expectedCount)
	}, timeout, DefaultHealthCheckInterval,
		"expected %d documents in index %q, timeout after %v",
		expectedCount, indexName, timeout)
}
