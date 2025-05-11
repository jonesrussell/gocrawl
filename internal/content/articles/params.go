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
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"articleIndexName"`
}

// ContentServiceResult contains the article service
type ContentServiceResult struct {
	fx.Out

	Service Interface `group:"services"`
}

// ProcessorParams contains dependencies for creating the article processor
type ProcessorParams struct {
	fx.In

	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string `name:"articleIndexName"`
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}

// ProcessorResult contains the article processor
type ProcessorResult struct {
	fx.Out

	Processor content.Processor `name:"articleProcessor"`
}
