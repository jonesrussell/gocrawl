package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Define a constant for the default scroll duration
const defaultScrollDuration = 5 * time.Minute

// getScrollDuration converts ScrollDuration string to time.Duration
func (s *ElasticsearchStorage) getScrollDuration() time.Duration {
	duration, err := time.ParseDuration(s.opts.ScrollDuration)
	if err != nil {
		// Default to the constant if parsing fails
		return defaultScrollDuration
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
		if decodeErr := json.NewDecoder(searchRes.Body).Decode(&e); decodeErr != nil {
			return nil, fmt.Errorf("error parsing error response: %w", decodeErr)
		}
		return nil, fmt.Errorf("elasticsearch error: %v", e["error"])
	}

	return searchRes, nil
}

// handleScrollRequests processes the scroll requests
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

		if fetchErr := s.fetchNextScrollBatch(scrollID, resultChan); fetchErr != nil {
			s.Logger.Error("Failed to get next scroll batch", "error", fetchErr)
			return
		}

		// Check for context cancellation
		if ctx.Err() != nil {
			s.Logger.Info("Context cancelled, stopping scroll")
			return
		}
	}
}

// fetchNextScrollBatch fetches the next batch of results from Elasticsearch
func (s *ElasticsearchStorage) fetchNextScrollBatch(
	scrollID string,
	resultChan chan<- map[string]interface{},
) error {
	searchRes, err := s.ESClient.Scroll(
		s.ESClient.Scroll.WithScrollID(scrollID),
		s.ESClient.Scroll.WithScroll(s.getScrollDuration()),
	)
	if err != nil {
		return err
	}

	if searchRes.IsError() {
		return s.handleScrollError(searchRes)
	}

	// Process the hits from the response
	return s.processScrollHits(searchRes, resultChan)
}

// handleScrollError handles errors from the scroll response
func (s *ElasticsearchStorage) handleScrollError(searchRes *esapi.Response) error {
	var e map[string]interface{}
	if decodeErr := json.NewDecoder(searchRes.Body).Decode(&e); decodeErr != nil {
		return decodeErr
	}

	// Check for specific error type
	if errMap, ok := e["error"].(map[string]interface{}); ok {
		if errType, ok := errMap["type"]; ok {
			if errType == "search_phase_execution_exception" {
				s.Logger.Info("Reached end of scroll results")
				return nil
			}
		}
	}

	s.Logger.Error("Unexpected scroll error", "error", e["error"])
	return fmt.Errorf("unexpected scroll error: %v", e["error"])
}

// processScrollHits processes the hits from the scroll response
func (s *ElasticsearchStorage) processScrollHits(
	searchRes *esapi.Response,
	resultChan chan<- map[string]interface{},
) error {
	var result map[string]interface{}
	if err := json.NewDecoder(searchRes.Body).Decode(&result); err != nil {
		return fmt.Errorf("error parsing scroll response: %w", err)
	}

	hits, err := s.getHitsFromResult(result)
	if err != nil {
		return err
	}

	s.ProcessHits(hits, resultChan)

	return nil
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

		// Send the source directly to the channel
		resultChan <- source
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

	s.ProcessHits(hits, resultChan)

	scrollID, ok := result["_scroll_id"].(string)
	if !ok {
		return "", ErrInvalidScrollID
	}

	return scrollID, nil
}

// ... rest of scroll-related methods ...
