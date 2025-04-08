// Package priority provides priority-related configuration types and functions.
package priority

import "fmt"

// Config holds priority-related configuration settings.
type Config struct {
	// Default is the default priority for sources
	Default int `yaml:"default"`
	// Rules define priority rules for specific sources
	Rules []Rule `yaml:"rules"`
}

// Rule defines a rule for setting source priority.
type Rule struct {
	// Pattern is a regex pattern to match source names
	Pattern string `yaml:"pattern"`
	// Priority is the priority value to set
	Priority int `yaml:"priority"`
}

// New creates a new priority configuration with default values.
func New() *Config {
	return &Config{
		Default: 0,
		Rules:   []Rule{},
	}
}

// Option is a function that configures a priority configuration.
type Option func(*Config)

// WithDefault sets the default priority.
func WithDefault(priority int) Option {
	return func(c *Config) {
		c.Default = priority
	}
}

// WithRule adds a rule to the configuration.
func WithRule(pattern string, priority int) Option {
	return func(c *Config) {
		c.Rules = append(c.Rules, Rule{
			Pattern:  pattern,
			Priority: priority,
		})
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	for _, rule := range c.Rules {
		if rule.Pattern == "" {
			return fmt.Errorf("rule pattern cannot be empty")
		}
	}

	return nil
}
