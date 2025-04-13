package types

import "errors"

// Rule represents a crawling rule.
type Rule struct {
	// Pattern is the URL pattern to match
	Pattern string `yaml:"pattern"`
	// Action is the action to take when the pattern matches
	Action string `yaml:"action"`
	// Priority is the priority of the rule
	Priority int `yaml:"priority"`
}

// Rules is a collection of crawling rules.
type Rules []Rule

// Validate validates the crawling rules.
func (r Rules) Validate() error {
	for i, rule := range r {
		if rule.Pattern == "" {
			return errors.New("pattern is required")
		}
		if rule.Action == "" {
			return errors.New("action is required")
		}
		if rule.Priority < 0 {
			return errors.New("priority must be non-negative")
		}
		// Check for duplicate patterns
		for j := i + 1; j < len(r); j++ {
			if rule.Pattern == r[j].Pattern {
				return errors.New("duplicate pattern found")
			}
		}
	}
	return nil
}
