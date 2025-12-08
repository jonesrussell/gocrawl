// Package generator provides tools for generating CSS selector configurations
// for news sources.
package generator

// SelectorCandidate represents a discovered CSS selector with confidence scoring.
type SelectorCandidate struct {
	// Field is the name of the field this selector targets (e.g., "title", "body")
	Field string
	// Selectors is a list of CSS selectors that match this field
	Selectors []string
	// Confidence is a score from 0.0 to 1.0 indicating how confident we are
	// that this selector correctly identifies the target field
	Confidence float64
	// SampleText contains the first 100 characters of extracted text for verification
	SampleText string
}

// DiscoveryResult holds all discovered selectors for a source.
type DiscoveryResult struct {
	// Title selector candidates
	Title SelectorCandidate
	// Body selector candidates
	Body SelectorCandidate
	// Author selector candidates
	Author SelectorCandidate
	// PublishedTime selector candidates
	PublishedTime SelectorCandidate
	// Image selector candidates
	Image SelectorCandidate
	// Link selector candidates (for article discovery)
	Link SelectorCandidate
	// Category selector candidates
	Category SelectorCandidate
	// Exclusions is a list of CSS selectors for elements to exclude
	Exclusions []string
}
