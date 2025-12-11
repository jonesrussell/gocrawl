// Package apiclient provides HTTP client functionality for interacting with the gosources API.
package apiclient

import "time"

// APISource represents a source as returned by the gosources API.
type APISource struct {
	ID           string       `json:"id,omitempty"`
	Name         string       `json:"name"`
	URL          string       `json:"url"`
	ArticleIndex string       `json:"article_index"`
	PageIndex    string       `json:"page_index"`
	RateLimit    string       `json:"rate_limit,omitempty"`
	MaxDepth     int          `json:"max_depth,omitempty"`
	Time         []string     `json:"time,omitempty"`
	Enabled      bool         `json:"enabled"`
	CityName     string       `json:"city_name,omitempty"`
	GroupID      string       `json:"group_id,omitempty"`
	Selectors    APISelectors `json:"selectors"`
	CreatedAt    *time.Time   `json:"created_at,omitempty"`
	UpdatedAt    *time.Time   `json:"updated_at,omitempty"`
}

// APISelectors represents the selectors structure in the API.
type APISelectors struct {
	Article APIArticleSelectors `json:"article"`
	List    APIListSelectors    `json:"list"`
	Page    APIPageSelectors    `json:"page"`
}

// APIArticleSelectors represents article selectors in the API.
type APIArticleSelectors struct {
	Container     string   `json:"container,omitempty"`
	Title         string   `json:"title,omitempty"`
	Body          string   `json:"body,omitempty"`
	Intro         string   `json:"intro,omitempty"`
	Link          string   `json:"link,omitempty"`
	Image         string   `json:"image,omitempty"`
	Byline        string   `json:"byline,omitempty"`
	PublishedTime string   `json:"published_time,omitempty"`
	TimeAgo       string   `json:"time_ago,omitempty"`
	Section       string   `json:"section,omitempty"`
	Category      string   `json:"category,omitempty"`
	ArticleID     string   `json:"article_id,omitempty"`
	JSONLD        string   `json:"json_ld,omitempty"`
	Keywords      string   `json:"keywords,omitempty"`
	Description   string   `json:"description,omitempty"`
	OGTitle       string   `json:"og_title,omitempty"`
	OGDescription string   `json:"og_description,omitempty"`
	OGImage       string   `json:"og_image,omitempty"`
	OgURL         string   `json:"og_url,omitempty"`
	OGType        string   `json:"og_type,omitempty"`
	OGSiteName    string   `json:"og_site_name,omitempty"`
	Canonical     string   `json:"canonical,omitempty"`
	Author        string   `json:"author,omitempty"`
	Exclude       []string `json:"exclude,omitempty"`
}

// APIListSelectors represents list selectors in the API.
type APIListSelectors struct {
	Container       string   `json:"container,omitempty"`
	ArticleCards    string   `json:"article_cards,omitempty"`
	ArticleList     string   `json:"article_list,omitempty"`
	ExcludeFromList []string `json:"exclude_from_list,omitempty"`
}

// APIPageSelectors represents page selectors in the API.
type APIPageSelectors struct {
	Container     string   `json:"container,omitempty"`
	Title         string   `json:"title,omitempty"`
	Content       string   `json:"content,omitempty"`
	Description   string   `json:"description,omitempty"`
	Keywords      string   `json:"keywords,omitempty"`
	OGTitle       string   `json:"og_title,omitempty"`
	OGDescription string   `json:"og_description,omitempty"`
	OGImage       string   `json:"og_image,omitempty"`
	OgURL         string   `json:"og_url,omitempty"`
	Canonical     string   `json:"canonical,omitempty"`
	Exclude       []string `json:"exclude,omitempty"`
}

// ListSourcesResponse represents the response from the list sources API.
type ListSourcesResponse struct {
	Sources []APISource `json:"sources"`
	Count   int         `json:"count"`
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
