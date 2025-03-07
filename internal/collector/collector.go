package collector

import (
	"fmt"
	"net/url"
)

// New creates a new collector instance
func New(p Params) (Result, error) {
	if err := ValidateParams(p); err != nil {
		return Result{}, err
	}

	// Create collector configuration
	cfg, err := NewConfig(p)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create collector config: %w", err)
	}

	if err := cfg.ValidateConfig(); err != nil {
		return Result{}, fmt.Errorf("invalid collector config: %w", err)
	}

	// Create collector setup
	setup := NewSetup(cfg)
	if err := setup.ValidateURL(); err != nil {
		return Result{}, err
	}

	// Parse URL to get domain
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return Result{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create base collector
	c := setup.CreateBaseCollector(parsedURL.Hostname())

	// Configure collector settings
	if err := setup.ConfigureCollector(c); err != nil {
		return Result{}, fmt.Errorf("failed to configure collector: %w", err)
	}

	// Create completion channel and handlers
	done := make(chan struct{})
	handlers := NewHandlers(cfg, done)
	handlers.ConfigureHandlers(c)

	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	return Result{Collector: c, Done: done}, nil
}
