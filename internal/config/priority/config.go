// Package priority provides priority-related configuration types and functions.
package priority

import "fmt"

// Config represents priority-specific configuration settings.
type Config struct {
	// DefaultPriority is the default priority for new items
	DefaultPriority int `yaml:"default_priority"`
	// MaxPriority is the maximum allowed priority
	MaxPriority int `yaml:"max_priority"`
	// MinPriority is the minimum allowed priority
	MinPriority int `yaml:"min_priority"`
	// PriorityIncrement is the amount to increment priority by
	PriorityIncrement int `yaml:"priority_increment"`
	// PriorityDecrement is the amount to decrement priority by
	PriorityDecrement int `yaml:"priority_decrement"`
	// Rules define priority rules for specific sources
	Rules []Rule `yaml:"rules"`
}

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		DefaultPriority:   5,
		MaxPriority:       10,
		MinPriority:       1,
		PriorityIncrement: 1,
		PriorityDecrement: 1,
	}
}

// Rule defines a rule for setting source priority.
type Rule struct {
	// Pattern is a regex pattern to match source names
	Pattern string `yaml:"pattern"`
	// Priority is the priority value to set
	Priority int `yaml:"priority"`
}

// Option is a function that configures a priority configuration.
type Option func(*Config)

// WithDefault sets the default priority.
func WithDefault(priority int) Option {
	return func(c *Config) {
		c.DefaultPriority = priority
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
