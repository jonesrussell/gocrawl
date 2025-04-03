// Package article provides functionality for processing and managing article content
// from web pages. It includes services for article extraction, processing, and storage,
// with support for configurable selectors and multiple content sources.
package article

import (
	"context"
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
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
// - Config: For application configuration
// - Sources: The content source interface
type ProcessorParams struct {
	fx.In

	Logger      logger.Interface
	Storage     types.Interface
	IndexName   string               `name:"indexName"`
	ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
	Service     Interface
	Config      config.Interface
	Sources     sources.Interface
	Source      string `name:"sourceName"`
}

// ServiceParams contains the dependencies required to create an article service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
// - Config: For application configuration
// - Storage: For article persistence
// - Source: The content source name
// - IndexName: The Elasticsearch index name for articles
// - Sources: The content source interface
type ServiceParams struct {
	fx.In

	Logger    logger.Interface
	Config    config.Interface
	Sources   sources.Interface
	Storage   types.Interface
	Source    string `name:"sourceName"`
	IndexName string `name:"indexName"`
}

// Manager handles article management operations.
type Manager struct {
	logger  logger.Interface
	storage types.Interface
	config  config.Interface
}

// Params holds the dependencies required for the article manager.
type Params struct {
	fx.In
	Context context.Context `name:"articleContext"`
	Config  config.Interface
	Logger  logger.Interface
	Storage types.Interface
}

// Module provides the article module and its dependencies.
var Module = fx.Module("article",
	fx.Provide(
		// Provide the article service
		func(p ServiceParams) (Interface, error) {
			// Get source configuration
			source, err := p.Sources.FindByName(p.Source)
			if err != nil {
				return nil, fmt.Errorf("failed to find source %s: %w", p.Source, err)
			}

			// Convert source selectors to article selectors
			selectors := config.ArticleSelectors{
				Container:     source.Selectors.Article.Container,
				Title:         source.Selectors.Article.Title,
				Body:          source.Selectors.Article.Body,
				Intro:         source.Selectors.Article.Intro,
				Byline:        source.Selectors.Article.Byline,
				PublishedTime: source.Selectors.Article.PublishedTime,
				TimeAgo:       source.Selectors.Article.TimeAgo,
				JSONLD:        source.Selectors.Article.JSONLD,
				Section:       source.Selectors.Article.Section,
				Keywords:      source.Selectors.Article.Keywords,
				Description:   source.Selectors.Article.Description,
				OGTitle:       source.Selectors.Article.OGTitle,
				OGDescription: source.Selectors.Article.OGDescription,
				OGImage:       source.Selectors.Article.OGImage,
				OgURL:         source.Selectors.Article.OgURL,
				Canonical:     source.Selectors.Article.Canonical,
				WordCount:     source.Selectors.Article.WordCount,
				PublishDate:   source.Selectors.Article.PublishDate,
				Category:      source.Selectors.Article.Category,
				Tags:          source.Selectors.Article.Tags,
				Author:        source.Selectors.Article.Author,
				BylineName:    source.Selectors.Article.BylineName,
			}

			// Use default selectors if source selectors are empty
			if isEmptySelectors(selectors) {
				selectors = config.DefaultArticleSelectors()
			}

			// Create service with selectors
			service := &articleService{
				storage:   p.Storage,
				log:       p.Logger,
				selectors: selectors,
			}
			p.Logger.Debug("Created article service", "type", fmt.Sprintf("%T", service))
			return service, nil
		},
		// Provide the article processor
		fx.Annotate(
			func(p ProcessorParams) common.Processor {
				processor := NewArticleProcessor(p)
				if processor == nil {
					panic("failed to create article processor")
				}
				return processor
			},
			fx.ResultTags(`name:"articleProcessor"`),
		),
		func(
			cfg config.Interface,
			log logger.Interface,
			storage types.Interface,
		) (*Manager, error) {
			return NewArticleManager(log, storage, cfg)
		},
	),
)

// articleService provides article management functionality.
type articleService struct {
	storage   types.Interface
	log       logger.Interface
	selectors config.ArticleSelectors
}

// ExtractArticle extracts article data from an HTML element.
func (s *articleService) ExtractArticle(e *colly.HTMLElement) *models.Article {
	// TODO: Implement article extraction
	return nil
}

// Process processes an article.
func (s *articleService) Process(article *models.Article) error {
	// TODO: Implement article processing
	return nil
}

// ProcessJob processes a job and its items.
func (s *articleService) ProcessJob(ctx context.Context, job *common.Job) {
	// TODO: Implement job processing
}

// ProcessHTML processes HTML content from a source.
func (s *articleService) ProcessHTML(e *colly.HTMLElement) error {
	// TODO: Implement HTML processing
	return nil
}

// GetMetrics returns the current processing metrics.
func (s *articleService) GetMetrics() *common.Metrics {
	// TODO: Implement metrics collection
	return &common.Metrics{}
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

// NewArticleManager creates a new article manager.
func NewArticleManager(log logger.Interface, storage types.Interface, cfg config.Interface) (*Manager, error) {
	return &Manager{
		logger:  log,
		storage: storage,
		config:  cfg,
	}, nil
}
