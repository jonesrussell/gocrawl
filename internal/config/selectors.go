package config

// ArticleSelectors defines the CSS selectors for article elements
type ArticleSelectors struct {
	Container     string `yaml:"container"`      // Container element for the article
	Title         string `yaml:"title"`          // Article title
	Body          string `yaml:"body"`           // Main article content
	Intro         string `yaml:"intro"`          // Article introduction/summary
	Byline        string `yaml:"byline"`         // Author byline
	PublishedTime string `yaml:"published_time"` // Published time meta tag
	TimeAgo       string `yaml:"time_ago"`       // Time ago element
	JSONLD        string `yaml:"json_ld"`        // JSON-LD script tag
	Section       string `yaml:"section"`        // Article section meta tag
	Keywords      string `yaml:"keywords"`       // Keywords meta tag
	Description   string `yaml:"description"`    // Description meta tag
	OGTitle       string `yaml:"og_title"`       // OpenGraph title
	OGDescription string `yaml:"og_description"` // OpenGraph description
	OGImage       string `yaml:"og_image"`       // OpenGraph image
	OGURL         string `yaml:"og_url"`         // OpenGraph URL
	Canonical     string `yaml:"canonical"`      // Canonical URL link
	WordCount     string `yaml:"word_count"`     // Word count script
	PublishDate   string `yaml:"publish_date"`   // Publish date script
	Category      string `yaml:"category"`       // Category script
	Tags          string `yaml:"tags"`           // Tags script
	Author        string `yaml:"author"`         // Author script
	BylineName    string `yaml:"byline_name"`    // Byline name script
}

// DefaultArticleSelectors returns default selectors that work for most sites
func DefaultArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		Container:     "article, .article",
		Title:         "h1",
		Body:          "article, .article-content, .post-content, .entry-content",
		Intro:         ".article-intro, .post-intro, .entry-summary",
		Byline:        ".article-byline, .post-meta, .entry-meta",
		PublishedTime: "meta[property='article:published_time']",
		TimeAgo:       "time",
		JSONLD:        "script[type='application/ld+json']",
		Section:       "meta[property='article:section']",
		Keywords:      "meta[name='keywords']",
		Description:   "meta[name='description']",
		OGTitle:       "meta[property='og:title']",
		OGDescription: "meta[property='og:description']",
		OGImage:       "meta[property='og:image']",
		OGURL:         "meta[property='og:url']",
		Canonical:     "link[rel='canonical']",
	}
}
