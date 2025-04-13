package types

import "errors"

// SourceSelectors defines the CSS selectors for extracting content.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
}

// Validate validates the source selectors.
func (s *SourceSelectors) Validate() error {
	return s.Article.Validate()
}

// ArticleSelectors defines the CSS selectors for article content.
type ArticleSelectors struct {
	// Container is the selector for the article container
	Container string `yaml:"container"`
	// Title is the selector for the article title
	Title string `yaml:"title"`
	// Body is the selector for the article body
	Body string `yaml:"body"`
	// Intro is the selector for the article introduction
	Intro string `yaml:"intro"`
	// Byline is the selector for the article byline
	Byline string `yaml:"byline"`
	// PublishedTime is the selector for the article published time
	PublishedTime string `yaml:"published_time"`
	// TimeAgo is the selector for the relative time
	TimeAgo string `yaml:"time_ago"`
	// JSONLD is the selector for JSON-LD metadata
	JSONLD string `yaml:"json_ld"`
	// Description is the selector for the article description
	Description string `yaml:"description"`
	// Section is the selector for the article section
	Section string `yaml:"section"`
	// Keywords is the selector for article keywords
	Keywords string `yaml:"keywords"`
	// OGTitle is the selector for the Open Graph title
	OGTitle string `yaml:"og_title"`
	// OGDescription is the selector for the Open Graph description
	OGDescription string `yaml:"og_description"`
	// OGImage is the selector for the Open Graph image
	OGImage string `yaml:"og_image"`
	// OgURL is the selector for the Open Graph URL
	OgURL string `yaml:"og_url"`
	// Canonical is the selector for the canonical URL
	Canonical string `yaml:"canonical"`
	// WordCount is the selector for the word count
	WordCount string `yaml:"word_count"`
	// PublishDate is the selector for the publish date
	PublishDate string `yaml:"publish_date"`
	// Category is the selector for the article category
	Category string `yaml:"category"`
	// Tags is the selector for article tags
	Tags string `yaml:"tags"`
	// Author is the selector for the article author
	Author string `yaml:"author"`
	// BylineName is the selector for the byline name
	BylineName string `yaml:"byline_name"`
}

// Validate validates the article selectors.
func (s *ArticleSelectors) Validate() error {
	if s.Container == "" {
		return errors.New("container selector is required")
	}
	if s.Title == "" {
		return errors.New("title selector is required")
	}
	if s.Body == "" {
		return errors.New("body selector is required")
	}
	return nil
}

// Default returns default article selectors.
func (s *ArticleSelectors) Default() ArticleSelectors {
	return ArticleSelectors{
		Container:     "article",
		Title:         "h1",
		Body:          "article > div",
		Intro:         "p.lead",
		Byline:        ".byline",
		PublishedTime: "time[datetime]",
		TimeAgo:       "time.ago",
		JSONLD:        "script[type='application/ld+json']",
		Description:   "meta[name='description']",
		Section:       ".section",
		Keywords:      "meta[name='keywords']",
		OGTitle:       "meta[property='og:title']",
		OGDescription: "meta[property='og:description']",
		OGImage:       "meta[property='og:image']",
		OgURL:         "meta[property='og:url']",
		Canonical:     "link[rel='canonical']",
		WordCount:     ".word-count",
		PublishDate:   "time[pubdate]",
		Category:      ".category",
		Tags:          ".tags",
		Author:        ".author",
		BylineName:    ".byline-name",
	}
}
