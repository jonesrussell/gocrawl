// Package elasticsearch provides Elasticsearch-specific implementations of the indexing interfaces.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/indexing/client"
	"github.com/jonesrussell/gocrawl/internal/indexing/errors"
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
		return nil, errors.NewSearchError(index, query, "search", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewSearchError(index, query, "search", fmt.Errorf("status: %s", res.Status()))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing hits")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing hits list")
	}

	var docs []interface{}
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		if source, ok := hitMap["_source"]; ok {
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
		return 0, errors.NewSearchError(index, query, "count", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, errors.NewSearchError(index, query, "count", fmt.Errorf("status: %s", res.Status()))
	}

	var result struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
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
		return nil, errors.NewSearchError(index, aggs, "aggregate", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewSearchError(index, aggs, "aggregate", fmt.Errorf("status: %s", res.Status()))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	aggregations, ok := result["aggregations"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing aggregations")
	}

	sm.logger.Info("Aggregation completed", "index", index)
	return aggregations, nil
}
