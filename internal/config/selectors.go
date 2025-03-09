// Package config provides configuration management for the GoCrawl application.
// This file specifically handles the configuration of CSS selectors used for
// extracting content from web pages.
package config

// ArticleSelectors defines the CSS selectors for article elements.
// It provides a comprehensive set of selectors for extracting various
// components of an article from a web page, including metadata and content.
type ArticleSelectors struct {
	// Container is the main wrapper element for the article content
	Container string `yaml:"container"`
	// Title is the selector for the article's main headline
	Title string `yaml:"title"`
	// Body is the selector for the article's main content
	Body string `yaml:"body"`
	// Intro is the selector for the article's introduction or summary
	Intro string `yaml:"intro"`
	// Byline is the selector for the article's author information
	Byline string `yaml:"byline"`
	// PublishedTime is the selector for the article's publication timestamp
	PublishedTime string `yaml:"published_time"`
	// TimeAgo is the selector for relative time information (e.g., "2 hours ago")
	TimeAgo string `yaml:"time_ago"`
	// JSONLD is the selector for JSON-LD structured data script tags
	JSONLD string `yaml:"json_ld"`
	// Section is the selector for the article's section or category
	Section string `yaml:"section"`
	// Keywords is the selector for article keywords or tags
	Keywords string `yaml:"keywords"`
	// Description is the selector for the article's meta description
	Description string `yaml:"description"`
	// OGTitle is the selector for OpenGraph title metadata
	OGTitle string `yaml:"og_title"`
	// OGDescription is the selector for OpenGraph description metadata
	OGDescription string `yaml:"og_description"`
	// OGImage is the selector for OpenGraph image metadata
	OGImage string `yaml:"og_image"`
	// OgURL is the selector for OpenGraph URL metadata
	OgURL string `yaml:"og_url"`
	// Canonical is the selector for the canonical URL link
	Canonical string `yaml:"canonical"`
	// WordCount is the selector for the article's word count
	WordCount string `yaml:"word_count"`
	// PublishDate is the selector for the article's publication date
	PublishDate string `yaml:"publish_date"`
	// Category is the selector for the article's category
	Category string `yaml:"category"`
	// Tags is the selector for the article's tags
	Tags string `yaml:"tags"`
	// Author is the selector for the article's author
	Author string `yaml:"author"`
	// BylineName is the selector for the author's name in the byline
	BylineName string `yaml:"byline_name"`
	// Exclude is a list of selectors for elements to remove from the content
	Exclude []string `yaml:"exclude"`
}

// DefaultArticleSelectors returns default selectors that work for most sites.
// These selectors are designed to work with common web page structures and
// follow standard HTML conventions. They can be used as a starting point
// and customized for specific sites as needed.
//
// Returns:
//   - ArticleSelectors: A set of default article selectors
func DefaultArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		// Container targets common article wrapper elements
		Container: "article, .article, [itemtype*='Article']",
		// Title targets the main headline
		Title: "h1",
		// Body targets common content wrapper elements
		Body: "article, [role='main'], .content, .article-content",
		// Intro targets common introduction elements
		Intro: ".article-intro, .post-intro, .entry-summary",
		// Byline targets common author information elements
		Byline: ".article-byline, .post-meta, .entry-meta",
		// PublishedTime targets standard metadata
		PublishedTime: "meta[property='article:published_time']",
		// TimeAgo targets time elements
		TimeAgo: "time",
		// JSONLD targets structured data
		JSONLD: "script[type='application/ld+json']",
		// Section targets category metadata
		Section: "meta[property='article:section']",
		// Keywords targets keyword metadata
		Keywords: "meta[name='keywords']",
		// Description targets description metadata
		Description: "meta[name='description']",
		// OGTitle targets OpenGraph title
		OGTitle: "meta[property='og:title']",
		// OGDescription targets OpenGraph description
		OGDescription: "meta[property='og:description']",
		// OGImage targets OpenGraph image
		OGImage: "meta[property='og:image']",
		// OgURL targets OpenGraph URL
		OgURL: "meta[property='og:url']",
		// Canonical targets canonical URL
		Canonical: "link[rel='canonical']",
	}
}
