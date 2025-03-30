// Package article provides functionality for processing and managing article content
// from web pages. It includes services for article extraction, processing, and storage,
// with support for configurable selectors and multiple content sources.
package article

import (
	"github.com/jonesrussell/gocrawl/internal/collector"
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

// Module provides article-related dependencies for the application.
// It provides:
// - Article service with configuration
// - Article processor with fx.Annotate for named dependencies
// The module uses fx.Provide to wire up dependencies and fx.Annotate
// to specify the processor as part of the "processors" group.
var Module = fx.Module("article",
	fx.Provide(
		NewServiceWithConfig,
		fx.Annotate(
			func(
				log common.Logger,
				service Interface,
				storage types.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) collector.Processor {
				return &ArticleProcessor{
					Logger:         log,
					ArticleService: service,
					Storage:        storage,
					IndexName:      params.IndexName,
					ArticleChan:    params.ArticleChan,
				}
			},
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// NewServiceWithConfig creates a new article service with configuration.
// It takes ServiceParams for dependency injection and returns an Interface.
// The function:
// 1. Retrieves article selectors from the configuration for the specified source
// 2. Falls back to default selectors if none are configured
// 3. Creates and returns a new service instance with the configured selectors
func NewServiceWithConfig(p ServiceParams) Interface {
	// Get the source configuration
	var selectors config.ArticleSelectors
	sources := p.Config.GetSources()
	for _, source := range sources {
		if source.Name == p.Source {
			// Use default selectors since we no longer have selectors in the Source struct
			selectors = config.ArticleSelectors{}
			break
		}
	}

	if isEmptySelectors(selectors) {
		p.Logger.Debug("Using default article selectors")
		selectors = config.DefaultArticleSelectors()
	} else {
		p.Logger.Debug("Using article selectors",
			"source", p.Source,
			"selectors", selectors)
	}

	return NewService(p.Logger, selectors, p.Storage, p.IndexName)
}

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
