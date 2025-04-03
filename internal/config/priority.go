// Package config provides configuration management for the application.
package config

// PriorityConfig represents the priority configuration for crawling.
type PriorityConfig struct {
	// Default is the default priority for URLs that don't match any rules
	Default int `yaml:"default"`
	// Rules is a list of priority rules
	Rules []PriorityRule `yaml:"rules"`
}

// PriorityRule defines a priority rule for URL matching
type PriorityRule struct {
	// Pattern is the regex pattern to match against URLs
	Pattern string `yaml:"pattern"`
	// Priority is the priority value for matching URLs
	Priority int `yaml:"priority"`
}
