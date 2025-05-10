// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/processor"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ContentServiceParams contains dependencies for creating the article service
type ContentServiceParams struct {
	Logger    logger.Interface
	Storage   types.Interface
	IndexName string
}

// ArticleServiceParams contains dependencies for creating the article service
type ArticleServiceParams struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"articleIndexName"`
}

// ArticleServiceResult contains the article service
type ArticleServiceResult struct {
	fx.Out

	Service Interface `group:"services"`
}

// ArticleProcessorParams contains dependencies for creating the article processor
type ArticleProcessorParams struct {
	fx.In

	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string `name:"articleIndexName"`
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}

// ArticleProcessorResult contains the article processor
type ArticleProcessorResult struct {
	fx.Out

	Processor content.Processor `name:"articleProcessor"`
}

// ProcessorParams contains dependencies for creating a processor
type ProcessorParams struct {
	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}
