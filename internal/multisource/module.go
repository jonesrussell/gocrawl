package multisource

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

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

// Start begins the crawling process for all sources
func (ms *MultiSource) Start(ctx context.Context) error {
	for _, source := range ms.Sources {
		if err := ms.crawlSource(ctx, source); err != nil {
			return fmt.Errorf("error crawling source %s: %w", source.Name, err)
		}
	}
	return nil
}

// crawlSource handles the crawling logic for a single source
func (ms *MultiSource) crawlSource(ctx context.Context, source SourceConfig) error {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout for the HTTP request
	}

	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", source.URL, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", source.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response from %s: %s", source.URL, resp.Status)
	}

	// Here you would parse the response body and extract the relevant content
	// For example, you could use an HTML parser to extract articles

	fmt.Printf("Successfully crawled %s and indexed to %s\n", source.URL, source.Index)
	// Implement indexing logic here (e.g., send data to Elasticsearch)

	return nil
}

// Stop halts the crawling process
func (ms *MultiSource) Stop() {
	// Logic to stop crawling if necessary
}

// Module provides the fx module for MultiSource
var Module = fx.Provide(NewMultiSource)
