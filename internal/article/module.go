// Package article provides functionality for processing and managing article content
// from web pages. It includes services for article extraction, processing, and storage,
// with support for configurable selectors and multiple content sources.
package article

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProcessorParams contains the dependencies required to create an ArticleProcessor.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
// - Storage: For article persistence
// - IndexName: The Elasticsearch index name for articles
// - ArticleChan: Channel for article processing
// - Service: The article service interface
type ProcessorParams struct {
	fx.In

	Logger      common.Logger
	Storage     types.Interface
	IndexName   string `name:"indexName"`
	ArticleChan chan *models.Article
	Service     Interface
}

// ServiceParams contains the dependencies required to create an article service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
// - Config: For application configuration
// - Storage: For article persistence
// - Source: The content source name
// - IndexName: The Elasticsearch index name for articles
type ServiceParams struct {
	fx.In

	Logger    common.Logger
	Config    config.Interface
	Storage   types.Interface
	Source    string `name:"sourceName"`
	IndexName string `name:"indexName"`
}

// Module provides the article module and its dependencies.
var Module = fx.Module("article",
	fx.Provide(
		// Provide the article service
		func(p ServiceParams) (Interface, error) {
			// Get source configuration
			sources := p.Config.GetSources()
			var selectors config.ArticleSelectors

			// Find selectors for the current source
			for _, source := range sources {
				if source.Name == p.Source {
					selectors = source.Selectors.Article
					break
				}
			}

			// Use default selectors if source selectors are empty
			if isEmptySelectors(selectors) {
				selectors = config.DefaultArticleSelectors()
			}

			// Create service with selectors
			service := NewService(p.Logger, selectors, p.Storage, p.IndexName)
			p.Logger.Debug("Created article service", "type", fmt.Sprintf("%T", service))
			return service, nil
		},
		// Provide the article processor
		fx.Annotate(
			func(p ProcessorParams) common.Processor {
				return NewArticleProcessor(p)
			},
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// isEmptySelectors checks if the article selectors are empty.
// It returns true if all selector fields are empty strings.
// This is used to determine if default selectors should be used.
func isEmptySelectors(s config.ArticleSelectors) bool {
	return s.Container == "" &&
		s.Title == "" &&
		s.Body == "" &&
		s.Intro == "" &&
		s.Byline == "" &&
		s.PublishedTime == "" &&
		s.TimeAgo == "" &&
		s.JSONLD == "" &&
		s.Section == "" &&
		s.Keywords == "" &&
		s.Description == "" &&
		s.OGTitle == "" &&
		s.OGDescription == "" &&
		s.OGImage == "" &&
		s.OgURL == "" &&
		s.Canonical == ""
}
