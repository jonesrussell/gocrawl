// Package elasticsearch provides Elasticsearch-specific implementations of the indexing interfaces.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/indexing/client"
	indexerrors "github.com/jonesrussell/gocrawl/internal/indexing/errors"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

var (
	errInvalidHits        = storage.ErrInvalidHits
	errInvalidHitsList    = storage.ErrInvalidHitsArray
	errInvalidAggregation = errors.New("invalid response format: missing aggregations")
)

// SearchManager implements the api.SearchManager interface using Elasticsearch.
type SearchManager struct {
	client *client.Client
	logger common.Logger
}

// NewSearchManager creates a new Elasticsearch-based search manager.
func NewSearchManager(client *client.Client, logger common.Logger) (api.SearchManager, error) {
	return &SearchManager{
		client: client,
		logger: logger,
	}, nil
}

// Search performs a search query.
func (sm *SearchManager) Search(ctx context.Context, index string, query interface{}) ([]interface{}, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, sm.client.Client())
	if err != nil {
		return nil, indexerrors.NewSearchError(index, query, "search", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, indexerrors.NewSearchError(index, query, "search", fmt.Errorf("status: %s", res.Status()))
	}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	hits, isMap := result["hits"].(map[string]interface{})
	if !isMap {
		return nil, errInvalidHits
	}

	hitsList, isList := hits["hits"].([]interface{})
	if !isList {
		return nil, errInvalidHitsList
	}

	var docs []interface{}
	for _, hit := range hitsList {
		hitMap, isMap := hit.(map[string]interface{})
		if !isMap {
			continue
		}
		if source, hasSource := hitMap["_source"]; hasSource {
			docs = append(docs, source)
		}
	}

	sm.logger.Info("Search completed", "index", index, "hits", len(docs))
	return docs, nil
}

// Count returns the number of documents matching a query.
func (sm *SearchManager) Count(ctx context.Context, index string, query interface{}) (int64, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.CountRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, sm.client.Client())
	if err != nil {
		return 0, indexerrors.NewSearchError(index, query, "count", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, indexerrors.NewSearchError(index, query, "count", fmt.Errorf("status: %s", res.Status()))
	}

	var result struct {
		Count int64 `json:"count"`
	}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return 0, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	sm.logger.Info("Count completed", "index", index, "count", result.Count)
	return result.Count, nil
}

// Aggregate performs an aggregation query.
func (sm *SearchManager) Aggregate(ctx context.Context, index string, aggs interface{}) (interface{}, error) {
	body, err := json.Marshal(map[string]interface{}{
		"size": 0,
		"aggs": aggs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal aggregation: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, sm.client.Client())
	if err != nil {
		return nil, indexerrors.NewSearchError(index, aggs, "aggregate", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, indexerrors.NewSearchError(index, aggs, "aggregate", fmt.Errorf("status: %s", res.Status()))
	}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	aggregations, isMap := result["aggregations"].(map[string]interface{})
	if !isMap {
		return nil, errInvalidAggregation
	}

	sm.logger.Info("Aggregation completed", "index", index)
	return aggregations, nil
}
