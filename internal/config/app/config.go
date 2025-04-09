package app

import (
	"errors"
	"fmt"
)

// Config represents application-specific configuration settings.
type Config struct {
	// Name is the name of the application
	Name string `yaml:"name"`
	// Version is the version of the application
	Version string `yaml:"version"`
	// Environment is the application environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Debug indicates whether debug mode is enabled
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

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Name:        "gocrawl",
		Version:     "1.0.0",
		Environment: "development",
		Debug:       false,
	}
}
