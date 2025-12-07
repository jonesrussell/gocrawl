// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/processor"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentServiceParams contains dependencies for creating the article service
type ContentServiceParams struct {
	Logger    logger.Interface
	Storage   types.Interface
	IndexName string
}

// ProcessorParams contains dependencies for creating the article processor
type ProcessorParams struct {
	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}
