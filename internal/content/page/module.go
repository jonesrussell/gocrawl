// Package page provides functionality for processing and managing web pages.
// It includes services for page extraction, processing, and storage, with support
// for different page types and formats.
package page

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProvidePageService creates a new page service
func ProvidePageService(p struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"pageIndexName"`
}) (struct {
	fx.Out

	Service Interface `group:"services"`
}, error) {
	service := NewContentService(p.Logger, p.Storage, p.IndexName)

	return struct {
		fx.Out
		Service Interface `group:"services"`
	}{
		Service: service,
	}, nil
}

// ProvidePageProcessor creates a new page processor
func ProvidePageProcessor(p struct {
	fx.In

	Logger      logger.Interface
	Service     Interface
	Validator   content.JobValidator
	Storage     types.Interface
	IndexName   string `name:"pageIndexName"`
	PageChannel chan *models.Page
}) (struct {
	fx.Out

	Processor content.Processor `name:"pageProcessor"`
}, error) {
	processor := NewPageProcessor(
		p.Logger,
		p.Service,
		p.Validator,
		p.Storage,
		p.IndexName,
		p.PageChannel,
	)

	return struct {
		fx.Out
		Processor content.Processor `name:"pageProcessor"`
	}{
		Processor: processor,
	}, nil
}

// Module provides the page module's dependencies.
var Module = fx.Module("page",
	fx.Provide(
		ProvidePageService,
		ProvidePageProcessor,
	),
)
