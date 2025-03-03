// Package article handles article-related operations
package article

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
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

func (s *Service) ExtractArticle(e *colly.HTMLElement) *models.Article {
	var jsonLD JSONLDArticle

	s.Logger.Debug("Extracting article",
		"component", "article/service",
		"url", e.Request.URL.String())

	// Extract metadata from JSON-LD first
	e.ForEach(s.Selectors.JSONLD, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			s.Logger.Debug("Failed to parse JSON-LD",
				"component", "article/service",
				"error", err)
		}
	})

	// Get a copy of the body element for manipulation
	bodyEl := e.DOM.Find(s.Selectors.Body)

	// Remove excluded elements from body
	for _, excludeSelector := range s.Selectors.Exclude {
		bodyEl.Find(excludeSelector).Remove()
	}

	// Extract body text after removing excluded elements
	body := bodyEl.Text()
	if body == "" && jsonLD.ArticleBody != "" {
		body = jsonLD.ArticleBody // Fallback to JSON-LD body
	}

	// Get intro/description with fallbacks
	intro := e.ChildText(s.Selectors.Intro)
	if intro == "" {
		intro = e.ChildAttr(s.Selectors.Description, "content")
	}
	if intro == "" && jsonLD.Description != "" {
		intro = jsonLD.Description
	}
	if intro != "" {
		body = intro + "\n\n" + body
	}

	// Extract title with fallbacks
	title := e.ChildText(s.Selectors.Title)
	if title == "" {
		title = e.ChildAttr(s.Selectors.OGTitle, "content")
	}
	if title == "" && jsonLD.Headline != "" {
		title = jsonLD.Headline
	}
	if title == "" && jsonLD.Name != "" {
		title = jsonLD.Name
	}

	// Extract author with fallbacks
	author := e.ChildText(s.Selectors.Byline)
	if author == "" {
		author = e.ChildText(s.Selectors.Author)
	}
	if author == "" && jsonLD.Author != "" {
		author = jsonLD.Author
	}
	author = s.CleanAuthor(author)

	// Extract byline name
	bylineName := e.ChildText(s.Selectors.BylineName)
	if bylineName == "" {
		bylineName = author // Use author as fallback
	}

	// Extract category and section
	category := e.ChildAttr(s.Selectors.Category, "content")
	if category == "" && jsonLD.Section != "" {
		category = jsonLD.Section
	}

	section := e.ChildAttr(s.Selectors.Section, "content")
	if section == "" && jsonLD.Section != "" {
		section = jsonLD.Section
	}

	// Extract word count with fallback to JSON-LD
	wordCount := 0
	if s.Selectors.WordCount != "" {
		wordCountStr := e.ChildText(s.Selectors.WordCount)
		if count, err := strconv.Atoi(wordCountStr); err == nil {
			wordCount = count
		}
	}
	if wordCount == 0 && jsonLD.WordCount > 0 {
		wordCount = jsonLD.WordCount
	}

	// Extract canonical URL with fallback to JSON-LD
	canonicalURL := e.ChildAttr(s.Selectors.Canonical, "href")
	if canonicalURL == "" && jsonLD.URL != "" {
		canonicalURL = jsonLD.URL
	}
	if canonicalURL == "" && jsonLD.MainEntityOfPage.ID != "" {
		canonicalURL = jsonLD.MainEntityOfPage.ID
	}

	// Extract image URL with fallback to JSON-LD
	ogImage := e.ChildAttr(s.Selectors.OGImage, "content")
	if ogImage == "" && jsonLD.Image != "" {
		ogImage = jsonLD.Image
	}

	// Create article with all available fields
	article := &models.Article{
		ID:            uuid.New().String(),
		Title:         title,
		Body:          body,
		Author:        author,
		BylineName:    bylineName,
		Source:        e.Request.URL.String(),
		Tags:          s.ExtractTags(e, jsonLD),
		Intro:         intro,
		Description:   e.ChildAttr(s.Selectors.Description, "content"),
		OgTitle:       e.ChildAttr(s.Selectors.OGTitle, "content"),
		OgDescription: e.ChildAttr(s.Selectors.OGDescription, "content"),
		OgImage:       ogImage,
		OgUrl:         e.ChildAttr(s.Selectors.OGURL, "content"),
		CanonicalUrl:  canonicalURL,
		WordCount:     wordCount,
		Category:      category,
		Section:       section,
		Keywords:      strings.Split(e.ChildAttr(s.Selectors.Keywords, "content"), ","),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Parse published date
	article.PublishedDate = s.ParsePublishedDate(e, jsonLD)

	// Skip empty articles
	if article.Title == "" && article.Body == "" {
		s.Logger.Debug("Skipping empty article",
			"component", "article/service",
			"url", article.Source)
		return nil
	}

	s.Logger.Debug("Extracted article",
		"component", "article/service",
		"id", article.ID,
		"title", article.Title,
		"url", article.Source,
		"date", article.PublishedDate,
		"author", article.Author,
		"tags", article.Tags,
		"wordCount", article.WordCount,
		"category", article.Category,
		"section", article.Section)

	return article
}

// CleanAuthor cleans up the author string
func (s *Service) CleanAuthor(author string) string {
	if idx := strings.Index(author, "    "); idx != -1 {
		author = strings.TrimSpace(author[:idx])
	}
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
	return removeDuplicates(tags)
}

func removeDuplicates(tags []string) []string {
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

// RemoveDuplicates removes duplicate tags from a slice
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
