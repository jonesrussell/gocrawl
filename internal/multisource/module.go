package multisource

import (
	"fmt"
	"os"

	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

// SourceConfig represents the configuration for a single source
type SourceConfig struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Index string `yaml:"index"`
}

// MultiSource manages multiple sources for crawling
type MultiSource struct {
	Sources []SourceConfig
}

// NewMultiSource creates a new MultiSource instance
func NewMultiSource() (*MultiSource, error) {
	var ms MultiSource
	data, err := os.ReadFile("sources.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read sources.yml: %w", err)
	}
	if err := yaml.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sources.yml: %w", err)
	}
	return &ms, nil
}

// Stop halts the crawling process
func (ms *MultiSource) Stop() {
	// Logic to stop crawling if necessary
}

// Module provides the fx module for MultiSource
var Module = fx.Provide(NewMultiSource)
