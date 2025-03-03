package content

import (
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
			NewProcessor,
			fx.As(new(collector.ContentProcessor)),
		),
	),
)

// Params holds the parameters for creating a content processor
type Params struct {
	fx.In

	Service Interface
	Storage storage.Interface
	Logger  logger.Interface
}

// Result holds the content processor
type Result struct {
	fx.Out

	Processor *Processor `group:"processors"`
}

// New creates a new content processor with dependencies
func New(p Params) Result {
	return Result{
		Processor: NewProcessor(p.Service, p.Storage, p.Logger),
	}
}
