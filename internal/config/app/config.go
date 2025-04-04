package app

import (
	"errors"
	"fmt"
)

// Config holds application-level configuration settings.
type Config struct {
	// Environment specifies the runtime environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Name is the application name
	Name string `yaml:"name"`
	// Version is the application version
	Version string `yaml:"version"`
	// Debug enables debug mode for additional logging
	Debug bool `yaml:"debug"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment must be specified")
	}

	switch c.Environment {
	case "development", "staging", "production":
		// Valid environment
	default:
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	if c.Name == "" {
		return errors.New("application name must be specified")
	}

	if c.Version == "" {
		return errors.New("application version must be specified")
	}

	return nil
}

// New creates a new application configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		Environment: "development",
		Name:        "gocrawl",
		Version:     "0.1.0",
		Debug:       false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures an application configuration.
type Option func(*Config)

// WithEnvironment sets the environment.
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
	}
}

// WithName sets the application name.
func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

// WithVersion sets the application version.
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithDebug sets the debug mode.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}
