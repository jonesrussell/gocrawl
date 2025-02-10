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

type Service struct {
	logger logger.Interface
}

func NewService(logger logger.Interface) *Service {
	return &Service{logger: logger}
}

func (s *Service) ExtractArticle(e *colly.HTMLElement) *models.Article {
	// Extract metadata from JSON-LD first
	var jsonLD struct {
		DateCreated   string   `json:"dateCreated"`
		DateModified  string   `json:"dateModified"`
		DatePublished string   `json:"datePublished"`
		Author        string   `json:"author"`
		Keywords      []string `json:"keywords"`
		Section       string   `json:"articleSection"`
	}

	e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
		if err := json.Unmarshal([]byte(el.Text), &jsonLD); err != nil {
			s.logger.Debug("Failed to parse JSON-LD", "error", err)
		}
	})

	// Clean up author (remove date)
	author := e.ChildText(".details-byline")
	if idx := strings.Index(author, "    "); idx != -1 {
		author = strings.TrimSpace(author[:idx])
	}

	// Create article with basic fields
	article := &models.Article{
		ID:     uuid.New().String(),
		Title:  e.ChildText("h1.details-title"),
		Body:   e.ChildText("#details-body"),
		Source: e.Request.URL.String(),
		Author: author,
		Tags:   make([]string, 0),
	}

	// Get intro/description
	if intro := e.ChildText(".details-intro"); intro != "" {
		article.Body = intro + "\n\n" + article.Body
	}

	// Parse published date - try multiple sources
	if timeStr := e.ChildAttr("meta[property='article:published_time']", "content"); timeStr != "" {
		s.logger.Debug("Found meta published time", "value", timeStr)
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			article.PublishedDate = t
		}
	} else if timeStr := e.ChildAttr("time.timeago", "datetime"); timeStr != "" {
		s.logger.Debug("Found timeago datetime", "value", timeStr)
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			article.PublishedDate = t
		}
	} else if jsonLD.DatePublished != "" {
		s.logger.Debug("Found JSON-LD published date", "value", jsonLD.DatePublished)
		if t, err := time.Parse(time.RFC3339, jsonLD.DatePublished); err == nil {
			article.PublishedDate = t
		}
	}

	// Add tags from multiple sources
	// 1. Article section from meta tag
	if section := e.ChildAttr("meta[property='article:section']", "content"); section != "" {
		s.logger.Debug("Found meta section", "value", section)
		article.Tags = append(article.Tags, section)
	}

	// 2. Keywords from meta tag
	if keywords := e.ChildAttr("meta[name='keywords']", "content"); keywords != "" {
		s.logger.Debug("Found meta keywords", "value", keywords)
		for _, tag := range strings.Split(keywords, "|") {
			if tag = strings.TrimSpace(tag); tag != "" {
				article.Tags = append(article.Tags, tag)
			}
		}
	}

	// 3. Add section from URL path
	if strings.Contains(article.Source, "/opp-beat/") {
		article.Tags = append(article.Tags, "OPP Beat")
	}

	// Remove duplicates from tags
	seen := make(map[string]bool)
	uniqueTags := make([]string, 0)
	for _, tag := range article.Tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	article.Tags = uniqueTags

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
