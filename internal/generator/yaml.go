// Package generator provides tools for generating CSS selector configurations
// for news sources.
package generator

import (
	"fmt"
	"net/url"
	"strings"
)

// GenerateSourceYAML generates a YAML configuration entry for a source.
func GenerateSourceYAML(
	sourceURL string,
	result DiscoveryResult,
) (string, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	var builder strings.Builder
	writeYAMLHeader(&builder, parsedURL, sourceURL)
	writeYAMLSelectors(&builder, result)
	writeYAMLExclusions(&builder, result)

	return builder.String(), nil
}

// writeYAMLHeader writes the header section of the YAML.
func writeYAMLHeader(builder *strings.Builder, parsedURL *url.URL, sourceURL string) {
	sourceName := generateSourceName(parsedURL.Hostname())
	articleIndex := generateIndexName(parsedURL.Hostname(), "articles")
	pageIndex := generateIndexName(parsedURL.Hostname(), "pages")

	builder.WriteString("  - name: \"")
	builder.WriteString(sourceName)
	builder.WriteString("\"\n")

	builder.WriteString("    url: \"")
	builder.WriteString(sourceURL)
	builder.WriteString("\"\n")

	builder.WriteString("    article_index: \"")
	builder.WriteString(articleIndex)
	builder.WriteString("\"\n")

	builder.WriteString("    page_index: \"")
	builder.WriteString(pageIndex)
	builder.WriteString("\"\n")

	builder.WriteString("    rate_limit: 1s\n")
	builder.WriteString("    max_depth: 2\n")
	builder.WriteString("    time:\n")
	builder.WriteString("      - \"11:45\"\n")
	builder.WriteString("      - \"23:45\"\n")

	builder.WriteString("    selectors:\n")
	builder.WriteString("      article:\n")
}

// writeYAMLSelectors writes all selector fields.
func writeYAMLSelectors(builder *strings.Builder, result DiscoveryResult) {
	writeSelectorField(builder, "title", result.Title)
	writeSelectorField(builder, "body", result.Body)
	writeSelectorField(builder, "author", result.Author)
	writeSelectorField(builder, "published_time", result.PublishedTime)
	writeSelectorField(builder, "image", result.Image)
	writeSelectorField(builder, "link", result.Link)
	writeSelectorField(builder, "category", result.Category)
}

// writeSelectorField writes a single selector field.
func writeSelectorField(builder *strings.Builder, fieldName string, candidate SelectorCandidate) {
	if len(candidate.Selectors) == 0 {
		return
	}

	builder.WriteString("        ")
	builder.WriteString(fieldName)
	builder.WriteString(": \"")
	builder.WriteString(strings.Join(candidate.Selectors, ", "))
	builder.WriteString("\"  # Confidence: ")
	fmt.Fprintf(builder, "%.2f", candidate.Confidence)
	builder.WriteString("\n")

	if candidate.SampleText != "" {
		builder.WriteString("        # Sample: \"")
		builder.WriteString(escapeYAMLString(candidate.SampleText))
		builder.WriteString("\"\n")
	}
}

// writeYAMLExclusions writes the exclusions section.
func writeYAMLExclusions(builder *strings.Builder, result DiscoveryResult) {
	if len(result.Exclusions) == 0 {
		return
	}

	builder.WriteString("        exclude: [\n")
	for _, excl := range result.Exclusions {
		builder.WriteString("          \"")
		builder.WriteString(excl)
		builder.WriteString("\",\n")
	}
	builder.WriteString("        ]\n")
}

// generateSourceName converts a hostname to a title case source name.
// Example: "www.example.com" -> "Example Com"
func generateSourceName(hostname string) string {
	// Remove www. prefix
	hostname = strings.TrimPrefix(hostname, "www.")
	hostname = strings.TrimPrefix(hostname, "www")

	// Split by dots
	parts := strings.Split(hostname, ".")
	if len(parts) == 0 {
		return hostname
	}

	// Take the main domain part (usually first or second)
	var mainPart string
	const minPartsForDomain = 2
	if len(parts) >= minPartsForDomain {
		// Take the second-to-last part (e.g., "example" from "example.com")
		mainPart = parts[len(parts)-2]
	} else {
		mainPart = parts[0]
	}

	// Convert to title case
	if mainPart == "" {
		return hostname
	}

	// Capitalize first letter and handle common cases
	mainPart = strings.ToUpper(mainPart[:1]) + strings.ToLower(mainPart[1:])

	// Handle common TLDs
	tld := ""
	if len(parts) > 1 {
		tld = parts[len(parts)-1]
	}

	// For common cases, return just the main part
	if tld == "com" || tld == "org" || tld == "net" {
		return mainPart
	}

	// Otherwise, return main part + TLD
	if tld != "" {
		return mainPart + " " + strings.ToUpper(tld)
	}

	return mainPart
}

// generateIndexName converts a hostname to a snake_case index name.
// Example: "example.com" -> "example_com_articles"
func generateIndexName(hostname, suffix string) string {
	// Remove www. prefix
	hostname = strings.TrimPrefix(hostname, "www.")
	hostname = strings.TrimPrefix(hostname, "www")

	// Replace dots and hyphens with underscores
	hostname = strings.ReplaceAll(hostname, ".", "_")
	hostname = strings.ReplaceAll(hostname, "-", "_")

	// Convert to lowercase
	hostname = strings.ToLower(hostname)

	// Remove trailing underscores
	hostname = strings.Trim(hostname, "_")

	return hostname + "_" + suffix
}

// escapeYAMLString escapes special characters in YAML strings.
func escapeYAMLString(s string) string {
	// Escape backslashes first to avoid double-escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Then escape other characters
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
