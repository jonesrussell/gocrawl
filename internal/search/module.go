package search

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the search module and its dependencies
var Module = fx.Module("search",
	fx.Provide(
		ProvideSearchService, // Function to provide the search service
	),
)

// ProvideSearchService initializes the search service
func ProvideSearchService(esClient *elasticsearch.Client, cfg *config.Config, log logger.Interface) (*Service, error) {
	// Initialize and return the search service
	return NewSearchService(esClient, cfg, log), nil
}
