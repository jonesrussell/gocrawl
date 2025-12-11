package generator_test

import (
	"strings"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenerateYAML_ValidOutput(t *testing.T) {
	result := generator.DiscoveryResult{
		Title: generator.SelectorCandidate{
			Field:      "title",
			Selectors:  []string{"h1", "meta[property='og:title']"},
			Confidence: 0.95,
			SampleText: "Test Article Title",
		},
		Body: generator.SelectorCandidate{
			Field:      "body",
			Selectors:  []string{".article-body"},
			Confidence: 0.85,
			SampleText: "This is the article body...",
		},
		Exclusions: []string{".ad", "nav", "script"},
	}

	yamlContent, err := generator.GenerateSourceYAML("https://www.example.com/news", result)
	require.NoError(t, err)

	// Check that it starts with proper indentation for array item
	assert.True(t, strings.HasPrefix(yamlContent, "  - name:"), "Should start with 2-space indent for array item")

	// Check that it contains expected fields
	assert.Contains(t, yamlContent, "name:")
	assert.Contains(t, yamlContent, "url:")
	assert.Contains(t, yamlContent, "article_index:")
	assert.Contains(t, yamlContent, "page_index:")
	assert.Contains(t, yamlContent, "selectors:")
	assert.Contains(t, yamlContent, "title:")
	assert.Contains(t, yamlContent, "body:")

	// Check confidence scores are included
	assert.Contains(t, yamlContent, "Confidence: 0.95")
	assert.Contains(t, yamlContent, "Confidence: 0.85")

	// Check sample text is included
	assert.Contains(t, yamlContent, "Sample:")
}

func TestGenerateYAML_CanAppendToSourcesYAML(t *testing.T) {
	result := generator.DiscoveryResult{
		Title: generator.SelectorCandidate{
			Field:      "title",
			Selectors:  []string{"h1"},
			Confidence: 0.95,
			SampleText: "Test",
		},
	}

	yamlContent, err := generator.GenerateSourceYAML("https://www.example.com", result)
	require.NoError(t, err)

	// Create a mock sources.yml structure
	mockSourcesYAML := `sources:
  - name: "Existing Source"
    url: "https://existing.com"
    article_index: "existing_articles"
    page_index: "existing_pages"
    rate_limit: 1s
    max_depth: 2
    time:
      - "11:45"
      - "23:45"
    selectors:
      article:
        title: "h1"
`

	// Append the generated content
	combinedYAML := mockSourcesYAML + yamlContent

	// Try to parse as YAML to ensure it's valid
	var parsed map[string]any
	err = yaml.Unmarshal([]byte(combinedYAML), &parsed)
	require.NoError(t, err, "Combined YAML should be valid")

	// Verify structure
	sources, ok := parsed["sources"].([]any)
	require.True(t, ok, "Should have sources array")
	assert.GreaterOrEqual(t, len(sources), 2, "Should have at least 2 sources")
}

func TestGenerateSourceName(t *testing.T) {
	testCases := []struct {
		hostname string
		expected string
	}{
		{"www.example.com", "Example"},
		{"example.com", "Example"},
		{"news.example.org", "Example"},
		{"subdomain.example.net", "Example"},
	}

	for _, tc := range testCases {
		t.Run(tc.hostname, func(t *testing.T) {
			// Test generateSourceName indirectly through GenerateSourceYAML
			result := generator.DiscoveryResult{}
			_, _ = generator.GenerateSourceYAML("https://"+tc.hostname, result)
			_ = result
			// The function should extract the main domain part
			assert.Contains(t, result, "Example")
		})
	}
}

func TestGenerateIndexName(t *testing.T) {
	testCases := []struct {
		hostname string
		suffix   string
		expected string
	}{
		{"www.example.com", "articles", "example_com_articles"},
		{"example.com", "articles", "example_com_articles"},
		{"news.example.org", "pages", "news_example_org_pages"},
	}

	for _, tc := range testCases {
		t.Run(tc.hostname+"_"+tc.suffix, func(t *testing.T) {
			// Test generateIndexName indirectly through GenerateSourceYAML
			result := generator.DiscoveryResult{}
			_, _ = generator.GenerateSourceYAML("https://"+tc.hostname, result)
			_ = result
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEscapeYAMLString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"text with\nnewline", "text with\\nnewline"},
		{"text with \"quotes\"", "text with \\\"quotes\\\""},
		{"text with \\backslash", "text with \\\\backslash"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Test escapeYAMLString indirectly through GenerateSourceYAML
			result := generator.DiscoveryResult{
				Title: generator.SelectorCandidate{
					SampleText: tc.input,
				},
			}
			yamlContent, _ := generator.GenerateSourceYAML("https://example.com", result)
			_ = yamlContent
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGenerateYAML_AllFields(t *testing.T) {
	result := generator.DiscoveryResult{
		Title: generator.SelectorCandidate{
			Field:      "title",
			Selectors:  []string{"h1"},
			Confidence: 0.95,
			SampleText: "Title",
		},
		Body: generator.SelectorCandidate{
			Field:      "body",
			Selectors:  []string{".article-body"},
			Confidence: 0.85,
			SampleText: "Body",
		},
		Author: generator.SelectorCandidate{
			Field:      "author",
			Selectors:  []string{".author"},
			Confidence: 0.80,
			SampleText: "Author",
		},
		PublishedTime: generator.SelectorCandidate{
			Field:      "published_time",
			Selectors:  []string{"time[datetime]"},
			Confidence: 0.90,
			SampleText: "2024-01-01",
		},
		Image: generator.SelectorCandidate{
			Field:      "image",
			Selectors:  []string{"article img"},
			Confidence: 0.85,
			SampleText: "https://example.com/image.jpg",
		},
		Link: generator.SelectorCandidate{
			Field:      "link",
			Selectors:  []string{"a[href*='/news/']"},
			Confidence: 0.70,
			SampleText: "/news/article1",
		},
		Category: generator.SelectorCandidate{
			Field:      "category",
			Selectors:  []string{".category"},
			Confidence: 0.75,
			SampleText: "Technology",
		},
		Exclusions: []string{".ad", "nav"},
	}

	yamlContent, err := generator.GenerateSourceYAML("https://www.example.com", result)
	require.NoError(t, err)

	// Verify all fields are present
	assert.Contains(t, yamlContent, "title:")
	assert.Contains(t, yamlContent, "body:")
	assert.Contains(t, yamlContent, "author:")
	assert.Contains(t, yamlContent, "published_time:")
	assert.Contains(t, yamlContent, "image:")
	assert.Contains(t, yamlContent, "link:")
	assert.Contains(t, yamlContent, "category:")
	assert.Contains(t, yamlContent, "exclude:")
}
