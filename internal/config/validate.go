// Package config provides configuration management for the GoCrawl application.
package config

import "errors"

// ValidateConfig validates the configuration values.
// It checks for required fields and valid value ranges.
//
// Parameters:
//   - cfg: The configuration to validate
//
// Returns:
//   - error: Any validation errors that occurred
func ValidateConfig(cfg Interface) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	crawler := cfg.GetCrawlerConfig()
	if crawler == nil {
		return errors.New("crawler config cannot be nil")
	}

	if crawler.MaxDepth < 0 {
		return errors.New("max depth cannot be negative")
	}

	if crawler.Parallelism < 0 {
		return errors.New("parallelism cannot be negative")
	}

	if crawler.RandomDelay < 0 {
		return errors.New("random delay cannot be negative")
	}

	es := cfg.GetElasticsearchConfig()
	if es == nil {
		return errors.New("elasticsearch config cannot be nil")
	}

	if len(es.Addresses) == 0 {
		return errors.New("elasticsearch addresses cannot be empty")
	}

	log := cfg.GetLogConfig()
	if log == nil {
		return errors.New("log config cannot be nil")
	}

	if log.Level == "" {
		return errors.New("log level cannot be empty")
	}

	return nil
}
