// Package article handles article-related operations
package article

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// Service defines the interface for article operations
type Service interface {
	ExtractArticle(e *colly.HTMLElement) *models.Article
}

// ArticleService implements the Service interface
type ArticleService struct {
	logger logger.Interface
}

// NewService creates a new ArticleService instance
func NewService(logger logger.Interface) Service {
	return &ArticleService{logger: logger}
}

type jsonLDArticle struct {
	DateCreated   string   `json:"dateCreated"`
	DateModified  string   `json:"dateModified"`
	DatePublished string   `json:"datePublished"`
	Author        string   `json:"author"`
	Keywords      []string `json:"keywords"`
	Section       string   `json:"articleSection"`
}

func (s *ArticleService) ExtractArticle(e *colly.HTMLElement) *models.Article {
	var jsonLD jsonLDArticle

	// Extract metadata from JSON-LD first
	e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			s.logger.Debug("Failed to parse JSON-LD", "error", err)
		}
	})

	// Clean up author (remove date)
	author := s.CleanAuthor(e.ChildText(".details-byline"))

	// Create article with basic fields
	article := &models.Article{
		ID:     uuid.New().String(),
		Title:  e.ChildText("h1.details-title"),
		Body:   e.ChildText("#details-body"),
		Source: e.Request.URL.String(),
		Author: author,
		Tags:   s.extractTags(e, jsonLD),
	}

	// Get intro/description
	if intro := e.ChildText(".details-intro"); intro != "" {
		article.Body = intro + "\n\n" + article.Body
	}

	// Parse published date
	article.PublishedDate = s.ParsePublishedDate(e, jsonLD)

	// Skip empty articles
	if article.Title == "" && article.Body == "" {
		s.logger.Debug("Skipping empty article", "url", article.Source)
		return nil
	}

	s.logger.Debug("Extracted article",
		"id", article.ID,
		"title", article.Title,
		"url", article.Source,
		"date", article.PublishedDate,
		"author", article.Author,
		"tags", article.Tags)

	return article
}

func (s *ArticleService) CleanAuthor(author string) string {
	if idx := strings.Index(author, "    "); idx != -1 {
		author = strings.TrimSpace(author[:idx])
	}
	return author
}

func (s *ArticleService) extractTags(e *colly.HTMLElement, jsonLD jsonLDArticle) []string {
	tags := make([]string, 0)

	// 1. JSON-LD keywords
	if len(jsonLD.Keywords) > 0 {
		s.logger.Debug("Found JSON-LD keywords", "values", jsonLD.Keywords)
		tags = append(tags, jsonLD.Keywords...)
	}

	// 2. JSON-LD section
	if jsonLD.Section != "" {
		s.logger.Debug("Found JSON-LD section", "value", jsonLD.Section)
		tags = append(tags, jsonLD.Section)
	}

	// 3. Article section from meta tag
	if section := e.ChildAttr("meta[property='article:section']", "content"); section != "" {
		s.logger.Debug("Found meta section", "value", section)
		tags = append(tags, section)
	}

	// 4. Keywords from meta tag
	if keywords := e.ChildAttr("meta[name='keywords']", "content"); keywords != "" {
		s.logger.Debug("Found meta keywords", "value", keywords)
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

func (s *ArticleService) ParsePublishedDate(e *colly.HTMLElement, jsonLD jsonLDArticle) time.Time {
	var publishedDate time.Time

	timeFormats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.2030000Z",
		"2006-01-02 15:04:05",
	}

	datesToTry := []string{
		jsonLD.DatePublished,
		jsonLD.DateModified,
		jsonLD.DateCreated,
		e.ChildAttr("meta[property='article:published_time']", "content"),
		e.ChildAttr("time.timeago", "datetime"),
		e.ChildText("time.timeago"),
	}

	for _, dateStr := range datesToTry {
		if dateStr == "" {
			continue
		}
		s.logger.Debug("Trying to parse date", "value", dateStr)
		for _, format := range timeFormats {
			if t, err := time.Parse(format, dateStr); err == nil {
				publishedDate = t
				s.logger.Debug("Successfully parsed date",
					"source", dateStr,
					"format", format,
					"result", t)
				break
			} else {
				s.logger.Debug("Failed to parse date",
					"source", dateStr,
					"format", format,
					"error", err)
			}
		}
		if !publishedDate.IsZero() {
			break
		}
	}

	return publishedDate
}
