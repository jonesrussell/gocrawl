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
type Interface interface {
	ExtractArticle(e *colly.HTMLElement) *models.Article
	ExtractTags(e *colly.HTMLElement, jsonLD JSONLDArticle) []string
	CleanAuthor(author string) string
	ParsePublishedDate(e *colly.HTMLElement, jsonLD JSONLDArticle) time.Time
}

// Service implements the Service interface
type Service struct {
	Logger logger.Interface
}

// NewService creates a new Service instance
func NewService(logger logger.Interface) Interface {
	return &Service{Logger: logger}
}

type JSONLDArticle struct {
	DateCreated   string   `json:"dateCreated"`
	DateModified  string   `json:"dateModified"`
	DatePublished string   `json:"datePublished"`
	Author        string   `json:"author"`
	Keywords      []string `json:"keywords"`
	Section       string   `json:"articleSection"`
}

const (
	DetailsBylineSelector = ".details-byline"
	DetailsTitleSelector  = "h1.details-title"
	DetailsBodySelector   = "#details-body"
	DetailsIntroSelector  = ".details-intro"
	MetaPublishedTime     = "meta[property='article:published_time']"
	TimeAgoSelector       = "time.timeago"
)

func (s *Service) ExtractArticle(e *colly.HTMLElement) *models.Article {
	var jsonLD JSONLDArticle

	// Extract metadata from JSON-LD first
	e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			s.Logger.Debug("Failed to parse JSON-LD", "error", err)
		}
	})

	// Clean up author (remove date)
	author := s.CleanAuthor(e.ChildText(DetailsBylineSelector))

	// Create article with basic fields
	article := &models.Article{
		ID:     uuid.New().String(),
		Title:  e.ChildText(DetailsTitleSelector),
		Body:   e.ChildText(DetailsBodySelector),
		Source: e.Request.URL.String(),
		Author: author,
		Tags:   s.ExtractTags(e, jsonLD),
	}

	// Get intro/description
	if intro := e.ChildText(DetailsIntroSelector); intro != "" {
		article.Body = intro + "\n\n" + article.Body
	}

	// Parse published date
	article.PublishedDate = s.ParsePublishedDate(e, jsonLD)

	// Skip empty articles
	if article.Title == "" && article.Body == "" {
		s.Logger.Debug("Skipping empty article", "url", article.Source)
		return nil
	}

	s.Logger.Debug("Extracted article",
		"id", article.ID,
		"title", article.Title,
		"url", article.Source,
		"date", article.PublishedDate,
		"author", article.Author,
		"tags", article.Tags)

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
	if section := e.ChildAttr("meta[property='article:section']", "content"); section != "" {
		s.Logger.Debug("Found meta section", "value", section)
		tags = append(tags, section)
	}

	// 4. Keywords from meta tag
	if keywords := e.ChildAttr("meta[name='keywords']", "content"); keywords != "" {
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
		e.ChildAttr(MetaPublishedTime, "content"),
		e.ChildAttr(TimeAgoSelector, "datetime"),
		e.ChildText(TimeAgoSelector),
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
