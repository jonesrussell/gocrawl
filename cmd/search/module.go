// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"errors"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the search command dependencies
var Module = fx.Module("search",
	// Core dependencies
	config.Module,
	logger.Module,
	storage.Module,
	api.Module,
	fx.Invoke(func(cfg config.Interface, log logger.Interface) error {
		// Validate Elasticsearch configuration
		esConfig := cfg.GetElasticsearchConfig()
		if esConfig == nil {
			return errors.New("elasticsearch configuration is required")
		}
		return nil
	}),
)
