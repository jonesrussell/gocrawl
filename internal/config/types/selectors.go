package types

import "errors"

// SourceSelectors defines the CSS selectors for extracting content.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
	// List contains selectors for article list pages
	List ListSelectors `yaml:"list"`
	// Page contains selectors for page-specific content
	Page PageSelectors `yaml:"page"`
}

// ListSelectors defines the CSS selectors for article list page extraction.
type ListSelectors struct {
	// Container is the selector for the list container
	Container string `yaml:"container"`
	// ArticleCards is the selector for article card elements
	ArticleCards string `yaml:"article_cards"`
	// ArticleList is the selector for article list elements
	ArticleList string `yaml:"article_list"`
	// ExcludeFromList are selectors to exclude from list extraction
	ExcludeFromList []string `yaml:"exclude_from_list"`
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
	// Link is the selector for article links (CRITICAL for article discovery)
	Link string `yaml:"link"`
	// Image is the selector for article images
	Image string `yaml:"image"`
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
	// OGType is the selector for the Open Graph type
	OGType string `yaml:"og_type"`
	// OGSiteName is the selector for the Open Graph site name
	OGSiteName string `yaml:"og_site_name"`
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
	// ArticleID is the selector for article identifiers
	ArticleID string `yaml:"article_id"`
	// Exclude are selectors for elements to exclude from content extraction
	Exclude []string `yaml:"exclude"`
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

// PageSelectors defines the CSS selectors for page content.
type PageSelectors struct {
	// Container is the selector for the page container
	Container string `yaml:"container"`
	// Title is the selector for the page title
	Title string `yaml:"title"`
	// Content is the selector for the page content/body
	Content string `yaml:"content"`
	// Description is the selector for the page description
	Description string `yaml:"description"`
	// Keywords is the selector for page keywords
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
	// Exclude are selectors for elements to exclude from content extraction
	Exclude []string `yaml:"exclude"`
}

// Default returns default page selectors.
func (s *PageSelectors) Default() PageSelectors {
	return PageSelectors{
		Container:     "main, article, body",
		Title:         "h1, title",
		Content:       "main, article, .content",
		Description:   "meta[name='description']",
		Keywords:      "meta[name='keywords']",
		OGTitle:       "meta[property='og:title']",
		OGDescription: "meta[property='og:description']",
		OGImage:       "meta[property='og:image']",
		OgURL:         "meta[property='og:url']",
		Canonical:     "link[rel='canonical']",
		Exclude: []string{
			// Default exclude patterns for clean content extraction
			"script, style, noscript",
			".ad, .advertisement, [class*='ad']",
			".header, .footer, nav",
			"button, form",
			".sidebar, .comments",
		},
	}
}

// Validate validates the page selectors (optional validation).
func (s *PageSelectors) Validate() error {
	// Page selectors are optional, so no strict validation required
	return nil
}
