package content

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the content service and processor
var Module = fx.Options(
	fx.Provide(
		// Provide the content service
		fx.Annotate(
			NewService,
			fx.As(new(Interface)),
		),
		// Provide the content processor
		fx.Annotate(
			func(p Params) *Processor {
				return NewProcessor(p.Service, p.Storage, p.Logger, p.IndexName)
			},
			fx.As(new(models.ContentProcessor)),
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// Params holds the parameters for creating a content processor
type Params struct {
	fx.In

	Service   Interface
	Storage   storage.Interface
	Logger    logger.Interface
	IndexName string `name:"contentIndex"`
}
