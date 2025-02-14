package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// getScrollDuration converts ScrollDuration string to time.Duration
func (s *ElasticsearchStorage) getScrollDuration() time.Duration {
	duration, err := time.ParseDuration(s.opts.ScrollDuration)
	if err != nil {
		// Default to 5 minutes if parsing fails
		return 5 * time.Minute
	}
	return duration
}

// ScrollSearch implements scroll API for handling large result sets
func (s *ElasticsearchStorage) ScrollSearch(
	ctx context.Context,
	index string,
	query map[string]interface{},
	batchSize int,
) (<-chan map[string]interface{}, error) {
	resultChan := make(chan map[string]interface{})

	searchRes, err := s.initializeScroll(ctx, index, query, batchSize)
	if err != nil {
		return nil, err
	}

	go s.handleScrollRequests(ctx, searchRes, resultChan)

	return resultChan, nil
}

func (s *ElasticsearchStorage) initializeScroll(
	ctx context.Context,
	index string,
	query map[string]interface{},
	batchSize int,
) (*esapi.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	searchRes, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(index),
		s.ESClient.Search.WithBody(&buf),
		s.ESClient.Search.WithScroll(s.getScrollDuration()),
		s.ESClient.Search.WithSize(batchSize),
	)
	if err != nil {
		return nil, fmt.Errorf("initial scroll failed: %w", err)
	}

	if searchRes.IsError() {
		defer searchRes.Body.Close()
		var e map[string]interface{}
		if err := json.NewDecoder(searchRes.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing error response: %w", err)
		}
		return nil, fmt.Errorf("elasticsearch error: %v", e["error"])
	}

	return searchRes, nil
}

func (s *ElasticsearchStorage) handleScrollRequests(
	ctx context.Context,
	searchRes *esapi.Response,
	resultChan chan<- map[string]interface{},
) {
	defer close(resultChan)
	defer searchRes.Body.Close()

	for {
		scrollID, err := s.HandleScrollResponse(ctx, searchRes, resultChan)
		if err != nil {
			s.Logger.Error("Error processing scroll response", "error", err)
			return
		}

		select {
		case <-ctx.Done():
			s.Logger.Info("Context cancelled, stopping scroll")
			return
		default:
			searchRes, err = s.ESClient.Scroll(
				s.ESClient.Scroll.WithContext(ctx),
				s.ESClient.Scroll.WithScrollID(scrollID),
				s.ESClient.Scroll.WithScroll(s.getScrollDuration()),
			)
			if err != nil {
				s.Logger.Error("Failed to get next scroll batch", "error", err)
				return
			}

			// Check if we've reached the end of the scroll
			if searchRes.IsError() {
				var e map[string]interface{}
				if err := json.NewDecoder(searchRes.Body).Decode(&e); err != nil {
					s.Logger.Error("Error parsing scroll error response", "error", err)
					return
				}
				if e["error"].(map[string]interface{})["type"] == "search_phase_execution_exception" {
					s.Logger.Info("Reached end of scroll results")
					return
				}
				s.Logger.Error("Unexpected scroll error", "error", e["error"])
				return
			}
		}
	}
}

func (s *ElasticsearchStorage) getHitsFromResult(
	result map[string]interface{},
) ([]interface{}, error) {
	hitsMap, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, ErrInvalidHits
	}

	hitsArr, ok := hitsMap["hits"].([]interface{})
	if !ok {
		return nil, ErrInvalidHitsArray
	}

	return hitsArr, nil
}

// ProcessHits extracts documents from hits array and sends them to resultChan
func (s *ElasticsearchStorage) ProcessHits(
	ctx context.Context,
	hits []interface{},
	resultChan chan<- map[string]interface{},
) {
	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		select {
		case <-ctx.Done():
			s.Logger.Info("Context cancelled while processing hits")
			return
		default:
			select {
			case <-ctx.Done():
				s.Logger.Info("Context cancelled while sending hit")
				return
			case resultChan <- source:
			}
		}
	}
}

// HandleScrollResponse processes a single scroll response
func (s *ElasticsearchStorage) HandleScrollResponse(
	ctx context.Context,
	res *esapi.Response,
	resultChan chan<- map[string]interface{},
) (string, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error parsing scroll response: %w", err)
	}

	hits, err := s.getHitsFromResult(result)
	if err != nil {
		return "", err
	}

	s.ProcessHits(ctx, hits, resultChan)

	scrollID, ok := result["_scroll_id"].(string)
	if !ok {
		return "", ErrInvalidScrollID
	}

	return scrollID, nil
}

// ... rest of scroll-related methods ...
