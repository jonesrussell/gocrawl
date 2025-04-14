// Package page provides functionality for processing and managing web pages.
// It includes services for page extraction, processing, and storage, with support
// for different page types and formats.
package page

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the page module's dependencies.
var Module = fx.Module("page",
	fx.Provide(
		// Provide the page service
		fx.Annotate(
			NewContentService,
			fx.As(new(Interface)),
		),
		// Provide the page processor
		fx.Annotate(
			func(p struct {
				fx.In
				Logger      logger.Interface
				Service     Interface
				Validator   jobtypes.JobValidator
				Storage     types.Interface
				IndexName   string `name:"pageIndexName"`
				PageChannel chan *models.Page
			}) *PageProcessor {
				return NewPageProcessor(ProcessorParams{
					Logger:      p.Logger,
					Service:     p.Service,
					Validator:   p.Validator,
					Storage:     p.Storage,
					IndexName:   p.IndexName,
					PageChannel: p.PageChannel,
				})
			},
			fx.ResultTags(`name:"pageProcessor"`),
			fx.As(new(common.Processor)),
		),
	),
)
