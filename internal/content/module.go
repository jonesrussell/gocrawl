// Package content provides functionality for processing and managing general web content.
// It includes services for content extraction, processing, and storage, with support
// for different content types and formats. This package is designed to handle
// non-article content that may be encountered during web crawling.
package content

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProcessorParams defines the parameters for creating a new ContentProcessor.
type ProcessorParams struct {
	Logger    logger.Interface
	Service   Interface
	Storage   types.Interface
	IndexName string
}

// Params defines the parameters required for creating a content service.
// It uses fx.In for dependency injection and includes:
// - Logger: For logging operations
type Params struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"contentIndexName"`
}

// Module provides the content module's dependencies.
var Module = fx.Module("content",
	fx.Provide(
		// Provide the content service
		fx.Annotate(
			NewContentService,
			fx.As(new(Interface)),
		),
		// Provide the content processor
		fx.Annotate(
			NewContentProcessor,
			fx.ResultTags(`group:"processors"`, `name:"contentProcessor"`),
			fx.As(new(common.Processor)),
		),
	),
)

// NewContentService creates a new content service with source-specific rules.
func NewContentService(
	logger logger.Interface,
	storage types.Interface,
	config config.Interface,
) Interface {
	service := NewService(logger, storage)

	// Get sources from config
	srcs := config.GetSources()
	if len(srcs) == 0 {
		logger.Warn("No sources configured, using default content rules")
		return service
	}

	// Add source-specific rules
	for _, src := range srcs {
		rules := ContentRules{
			ContentTypePatterns: contentTypePatterns,
			ExcludePatterns:     make([]string, 0),
			MetadataSelectors: map[string]string{
				"title":       src.Selectors.Article.Title,
				"description": src.Selectors.Article.Description,
				"keywords":    src.Selectors.Article.Keywords,
				"author":      src.Selectors.Article.Author,
			},
			ContentSelectors: map[string]string{
				"body": src.Selectors.Article.Body,
			},
		}

		// Convert source rules to exclude patterns
		for _, rule := range src.Rules {
			if rule.Action == "disallow" {
				rules.ExcludePatterns = append(rules.ExcludePatterns, rule.Pattern)
			}
		}

		service.(*Service).AddSourceRules(src.Name, rules)
	}

	return service
}
