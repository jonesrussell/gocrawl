package storage

import (
	"context"
	"fmt"
	"io"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ElasticsearchStorage implements the Interface using Elasticsearch.
type ElasticsearchStorage struct {
	ESClient *es.Client
	Logger   logger.Interface
	opts     Options
}

// TestConnection tests the connection to Elasticsearch.
func (s *ElasticsearchStorage) TestConnection(_ context.Context) error {
	res, err := s.ESClient.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}
