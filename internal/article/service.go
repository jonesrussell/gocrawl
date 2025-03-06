// Package article handles article-related operations
package article

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

const (
	DefaultBodySelector = "article, .article"
)

// Service defines the interface for article operations
type Interface interface {
	ExtractArticle(e *colly.HTMLElement) *models.Article
	ExtractTags(e *colly.HTMLElement, jsonLD JSONLDArticle) []string
	CleanAuthor(author string) string
	ParsePublishedDate(e *colly.HTMLElement, jsonLD JSONLDArticle) time.Time
}

// Service implements the Service interface
type Service struct {
	Logger    logger.Interface
	Selectors config.ArticleSelectors
}

// Ensure Service implements Interface
var _ Interface = (*Service)(nil)

// NewService creates a new Service instance
func NewService(logger logger.Interface, selectors config.ArticleSelectors) Interface {
	return &Service{
		Logger:    logger,
		Selectors: selectors,
	}
}

type JSONLDArticle struct {
	DateCreated      string   `json:"dateCreated"`
	DateModified     string   `json:"dateModified"`
	DatePublished    string   `json:"datePublished"`
	Author           string   `json:"author"`
	Keywords         []string `json:"keywords"`
	Section          string   `json:"articleSection"`
	WordCount        int      `json:"wordCount"`
	Description      string   `json:"description"`
	Headline         string   `json:"headline"`
	ArticleBody      string   `json:"articleBody"`
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	Image            string   `json:"image"`
	MainEntityOfPage struct {
		ID   string `json:"@id"`
		Type string `json:"@type"`
	} `json:"mainEntityOfPage"`
	Publisher struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Logo struct {
			URL string `json:"url"`
		} `json:"logo"`
	} `json:"publisher"`
}

// ExtractMetadata extracts metadata from the HTML element
func (s *Service) ExtractMetadata(e *colly.HTMLElement) *models.Article {
	article := &models.Article{
		ID:        uuid.New().String(),
		Source:    e.Request.URL.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract metadata
	article.Title = s.extractTitle(e)
	article.Description = s.extractDescription(e)
	article.PublishedDate = s.parsePublishedTime(e)
	article.Author = s.extractAuthor(e)
	article.Section = s.extractSection(e)
	article.CanonicalURL = s.extractCanonicalURL(e)

	return article
}

// ExtractContent extracts the main content from the HTML element
func (s *Service) ExtractContent(e *colly.HTMLElement, article *models.Article) {
	bodyEl := s.findArticleBody(e)
	if bodyEl == nil {
		s.Logger.Debug("No article body found", "url", article.Source)
		return
	}

	article.Body = s.cleanAndExtractText(bodyEl)
	article.WordCount = len(strings.Fields(article.Body))
}

// ExtractArticle extracts article data from the HTML element
func (s *Service) ExtractArticle(e *colly.HTMLElement) *models.Article {
	article := s.ExtractMetadata(e)
	s.ExtractContent(e, article)
	return article
}

func (s *Service) extractTitle(e *colly.HTMLElement) string {
	// Try OpenGraph title first
	if title := e.ChildAttr(`meta[property="og:title"]`, "content"); title != "" {
		return title
	}

	// Try article title
	if title := e.ChildText(s.Selectors.Title); title != "" {
		return title
	}

	// Fallback to page title
	return e.ChildText("title")
}

func (s *Service) extractDescription(e *colly.HTMLElement) string {
	// Try OpenGraph description
	if desc := e.ChildAttr(`meta[property="og:description"]`, "content"); desc != "" {
		return desc
	}

	// Try meta description
	return e.ChildAttr(s.Selectors.Description, "content")
}

func (s *Service) extractPublishedTime(e *colly.HTMLElement) string {
	// Try article published time
	if time := e.ChildAttr(s.Selectors.PublishedTime, "content"); time != "" {
		return time
	}

	// Try meta published time
	return e.ChildAttr(`meta[property="article:published_time"]`, "content")
}

func (s *Service) extractAuthor(e *colly.HTMLElement) string {
	// Try article author
	if author := e.ChildText(s.Selectors.Byline + " " + s.Selectors.Author); author != "" {
		return author
	}

	// Try meta author
	return e.ChildAttr(s.Selectors.Author, "content")
}

func (s *Service) extractSection(e *colly.HTMLElement) string {
	return e.ChildText(s.Selectors.Section)
}

func (s *Service) extractCanonicalURL(e *colly.HTMLElement) string {
	return e.ChildAttr("link[rel=canonical]", "href")
}

func (s *Service) findArticleBody(e *colly.HTMLElement) *goquery.Selection {
	bodySelector := s.Selectors.Body
	if bodySelector == "" {
		bodySelector = DefaultBodySelector
	}
	return e.DOM.Find(bodySelector).First()
}

func (s *Service) cleanAndExtractText(bodyEl *goquery.Selection) string {
	// Remove unwanted elements
	bodyEl.Find("script, style, noscript, iframe, form").Remove()
	return strings.TrimSpace(bodyEl.Text())
}

func (s *Service) parsePublishedTime(e *colly.HTMLElement) time.Time {
	timeStr := s.extractPublishedTime(e)
	if timeStr == "" {
		return time.Time{}
	}

	// Try parsing with RFC3339 format first
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t
	}

	// Try parsing with RFC3339Nano format
	t, err = time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		return t
	}

	// Return zero time if parsing fails
	return time.Time{}
}

// CleanAuthor cleans up the author string
func (s *Service) CleanAuthor(author string) string {
	if author == "" {
		return ""
	}
	// Remove any extra whitespace
	author = strings.TrimSpace(author)
	return author
}

func (s *Service) ExtractTags(e *colly.HTMLElement, jsonLD JSONLDArticle) []string {
	tags := make([]string, 0)

	// 1. JSON-LD keywords
	if len(jsonLD.Keywords) > 0 {
		s.Logger.Debug("Found JSON-LD keywords", "values", jsonLD.Keywords)
		tags = append(tags, jsonLD.Keywords...)
	}

	// 2. JSON-LD section
	if jsonLD.Section != "" {
		s.Logger.Debug("Found JSON-LD section", "value", jsonLD.Section)
		tags = append(tags, jsonLD.Section)
	}

	// 3. Article section from meta tag
	if section := e.ChildAttr(s.Selectors.Section, "content"); section != "" {
		s.Logger.Debug("Found meta section", "value", section)
		tags = append(tags, section)
	}

	// 4. Keywords from meta tag
	if keywords := e.ChildAttr(s.Selectors.Keywords, "content"); keywords != "" {
		s.Logger.Debug("Found meta keywords", "value", keywords)
		for _, tag := range strings.Split(keywords, "|") {
			if tag = strings.TrimSpace(tag); tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// 5. Add section from URL path
	if strings.Contains(e.Request.URL.String(), "/opp-beat/") {
		tags = append(tags, "OPP Beat")
	}

	// Remove duplicates from tags
	return RemoveDuplicates(tags)
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(tags []string) []string {
	seen := make(map[string]bool)
	uniqueTags := make([]string, 0)
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	return uniqueTags
}

func (s *Service) ParsePublishedDate(e *colly.HTMLElement, jsonLD JSONLDArticle) time.Time {
	datesToTry := []string{
		jsonLD.DatePublished,
		jsonLD.DateModified,
		jsonLD.DateCreated,
		e.ChildAttr(s.Selectors.PublishedTime, "content"),
		e.ChildAttr(s.Selectors.TimeAgo, "datetime"),
		e.ChildText(s.Selectors.TimeAgo),
	}

	return parseDate(datesToTry, s.Logger)
}

func parseDate(dates []string, logger logger.Interface) time.Time {
	var publishedDate time.Time
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.2030000Z",
		"2006-01-02 15:04:05",
	}

	for _, dateStr := range dates {
		if dateStr == "" {
			continue
		}
		logger.Debug("Trying to parse date", "value", dateStr)
		for _, format := range timeFormats {
			t, err := time.Parse(format, dateStr)
			if err == nil {
				publishedDate = t
				logger.Debug("Successfully parsed date",
					"source", dateStr,
					"format", format,
					"result", t)
				break
			}
			logger.Debug("Failed to parse date",
				"source", dateStr,
				"format", format,
				"error", err)
		}
		if !publishedDate.IsZero() {
			break
		}
	}

	if publishedDate.IsZero() {
		logger.Debug("No valid published date found", "dates", dates)
	}

	return publishedDate
}
