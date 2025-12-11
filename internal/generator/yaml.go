// Package generator provides tools for generating CSS selector configurations
// for news sources.
package generator

import (
	"fmt"
	"net/url"
	"strings"
)

// GenerateSourceYAML generates a YAML configuration entry for a source
// that can be appended to sources.yml.
func GenerateSourceYAML(
	sourceURL string,
	result DiscoveryResult,
) (string, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Generate source name from hostname
	sourceName := generateSourceName(parsedURL.Hostname())

	// Generate index names
	articleIndex := generateIndexName(parsedURL.Hostname(), "articles")
	pageIndex := generateIndexName(parsedURL.Hostname(), "pages")

	var builder strings.Builder

	// Start with array item indent (2 spaces)
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

	// Title
	if len(result.Title.Selectors) > 0 {
		builder.WriteString("        title: \"")
		builder.WriteString(strings.Join(result.Title.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Title.Confidence))
		builder.WriteString("\n")
		if result.Title.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Title.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Body
	if len(result.Body.Selectors) > 0 {
		builder.WriteString("        body: \"")
		builder.WriteString(strings.Join(result.Body.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Body.Confidence))
		builder.WriteString("\n")
		if result.Body.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Body.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Author
	if len(result.Author.Selectors) > 0 {
		builder.WriteString("        author: \"")
		builder.WriteString(strings.Join(result.Author.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Author.Confidence))
		builder.WriteString("\n")
		if result.Author.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Author.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Published time
	if len(result.PublishedTime.Selectors) > 0 {
		builder.WriteString("        published_time: \"")
		builder.WriteString(strings.Join(result.PublishedTime.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.PublishedTime.Confidence))
		builder.WriteString("\n")
		if result.PublishedTime.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.PublishedTime.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Image
	if len(result.Image.Selectors) > 0 {
		builder.WriteString("        image: \"")
		builder.WriteString(strings.Join(result.Image.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Image.Confidence))
		builder.WriteString("\n")
		if result.Image.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Image.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Link
	if len(result.Link.Selectors) > 0 {
		builder.WriteString("        link: \"")
		builder.WriteString(strings.Join(result.Link.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Link.Confidence))
		builder.WriteString("\n")
		if result.Link.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Link.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Category
	if len(result.Category.Selectors) > 0 {
		builder.WriteString("        category: \"")
		builder.WriteString(strings.Join(result.Category.Selectors, ", "))
		builder.WriteString("\"  # Confidence: ")
		builder.WriteString(fmt.Sprintf("%.2f", result.Category.Confidence))
		builder.WriteString("\n")
		if result.Category.SampleText != "" {
			builder.WriteString("        # Sample: \"")
			builder.WriteString(escapeYAMLString(result.Category.SampleText))
			builder.WriteString("\"\n")
		}
	}

	// Exclusions
	if len(result.Exclusions) > 0 {
		builder.WriteString("        exclude: [\n")
		for _, excl := range result.Exclusions {
			builder.WriteString("          \"")
			builder.WriteString(excl)
			builder.WriteString("\"")
			builder.WriteString(",\n")
		}
		builder.WriteString("        ]\n")
	}

	return builder.String(), nil
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
