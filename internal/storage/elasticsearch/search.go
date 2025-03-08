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
)

var (
	errInvalidHitsFormat     = errors.New("invalid response format: missing hits")
	errInvalidHitsListFormat = errors.New("invalid response format: missing hits list")
	errInvalidAggregations   = errors.New("invalid response format: missing aggregations")
	errFailedDecode          = errors.New("failed to decode response")
	errFailedMarshalQuery    = errors.New("failed to marshal query")
	errFailedMarshalAggs     = errors.New("failed to marshal aggregation")
	errInvalidStatus         = errors.New("invalid response status")
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
		return nil, fmt.Errorf("%w: %w", errFailedMarshalQuery, err)
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
		statusErr := errors.New(res.Status())
		wrappedErr := fmt.Errorf("%w: %w", errInvalidStatus, statusErr)
		return nil, indexerrors.NewSearchError(index, query, "search", wrappedErr)
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(res.Body).Decode(&result)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: %w", errFailedDecode, decodeErr)
	}

	hits, isMap := result["hits"].(map[string]interface{})
	if !isMap {
		return nil, errInvalidHitsFormat
	}

	hitsList, isList := hits["hits"].([]interface{})
	if !isList {
		return nil, errInvalidHitsListFormat
	}

	var docs []interface{}
	for _, hit := range hitsList {
		hitMap, isValidMap := hit.(map[string]interface{})
		if !isValidMap {
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
		return 0, fmt.Errorf("%w: %w", errFailedMarshalQuery, err)
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
		statusErr := errors.New(res.Status())
		wrappedErr := fmt.Errorf("%w: %w", errInvalidStatus, statusErr)
		return 0, indexerrors.NewSearchError(index, query, "count", wrappedErr)
	}

	var result struct {
		Count int64 `json:"count"`
	}
	decodeErr := json.NewDecoder(res.Body).Decode(&result)
	if decodeErr != nil {
		return 0, fmt.Errorf("%w: %w", errFailedDecode, decodeErr)
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
		return nil, fmt.Errorf("%w: %w", errFailedMarshalAggs, err)
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
		statusErr := errors.New(res.Status())
		wrappedErr := fmt.Errorf("%w: %w", errInvalidStatus, statusErr)
		return nil, indexerrors.NewSearchError(index, aggs, "aggregate", wrappedErr)
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(res.Body).Decode(&result)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: %w", errFailedDecode, decodeErr)
	}

	aggregations, isMap := result["aggregations"].(map[string]interface{})
	if !isMap {
		return nil, errInvalidAggregations
	}

	sm.logger.Info("Aggregation completed", "index", index)
	return aggregations, nil
}
