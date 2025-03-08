// Package elasticsearch provides Elasticsearch-specific implementations of the indexing interfaces.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/indexing/client"
	"github.com/jonesrussell/gocrawl/internal/indexing/errors"
)

// DocumentManager implements the api.DocumentManager interface using Elasticsearch.
type DocumentManager struct {
	client *client.Client
	logger common.Logger
}

// NewDocumentManager creates a new Elasticsearch-based document manager.
func NewDocumentManager(client *client.Client, logger common.Logger) (api.DocumentManager, error) {
	return &DocumentManager{
		client: client,
		logger: logger,
	}, nil
}

// Index indexes a document with the given ID.
func (dm *DocumentManager) Index(ctx context.Context, index string, id string, doc interface{}) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
	}

	res, err := req.Do(ctx, dm.client.Client())
	if err != nil {
		return errors.NewDocumentError(index, id, "index", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewDocumentError(index, id, "index", fmt.Errorf("status: %s", res.Status()))
	}

	dm.logger.Info("Indexed document", "index", index, "id", id)
	return nil
}

// Update updates an existing document.
func (dm *DocumentManager) Update(ctx context.Context, index string, id string, doc interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"doc": doc,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
	}

	res, err := req.Do(ctx, dm.client.Client())
	if err != nil {
		return errors.NewDocumentError(index, id, "update", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return errors.ErrDocumentNotFound
		}
		return errors.NewDocumentError(index, id, "update", fmt.Errorf("status: %s", res.Status()))
	}

	dm.logger.Info("Updated document", "index", index, "id", id)
	return nil
}

// Delete deletes a document.
func (dm *DocumentManager) Delete(ctx context.Context, index string, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
	}

	res, err := req.Do(ctx, dm.client.Client())
	if err != nil {
		return errors.NewDocumentError(index, id, "delete", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return errors.ErrDocumentNotFound
		}
		return errors.NewDocumentError(index, id, "delete", fmt.Errorf("status: %s", res.Status()))
	}

	dm.logger.Info("Deleted document", "index", index, "id", id)
	return nil
}

// Get retrieves a document by ID.
func (dm *DocumentManager) Get(ctx context.Context, index string, id string) (interface{}, error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}

	res, err := req.Do(ctx, dm.client.Client())
	if err != nil {
		return nil, errors.NewDocumentError(index, id, "get", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, errors.ErrDocumentNotFound
		}
		return nil, errors.NewDocumentError(index, id, "get", fmt.Errorf("status: %s", res.Status()))
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(res.Body).Decode(&result)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	return result["_source"], nil
}
