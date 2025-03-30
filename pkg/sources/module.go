// Package sources provides source management functionality for the application.
package sources

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/gocrawl/pkg/logger"
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	fx.Provide(
		NewSources,
	),
)

// NewSources creates a new sources instance.
func NewSources(p Params) (Interface, error) {
	if err := ValidateParams(p); err != nil {
		return nil, err
	}
	return &sources{
		config: p.Config,
		logger: p.Logger,
	}, nil
}

type sources struct {
	config Interface
	logger logger.Interface
}

func (s *sources) GetSources() ([]Config, error) {
	return s.config.GetSources()
}

func (s *sources) FindByName(name string) (*Config, error) {
	return s.config.FindByName(name)
}

func (s *sources) Validate(source *Config) error {
	return s.config.Validate(source)
}
