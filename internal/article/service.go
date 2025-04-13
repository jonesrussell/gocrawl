// Package article provides functionality for processing and managing articles.
package article

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

const (
	DefaultBodySelector = "article, .article"
)

// Service implements article processing operations.
type Service struct {
	// Logger for article operations
	Logger logger.Interface
	// SourceSelectors maps source names to their selectors
	SourceSelectors map[string]types.ArticleSelectors
	// DefaultSelectors is used when no source-specific selectors are found
	DefaultSelectors types.ArticleSelectors
	// Storage for article persistence
	Storage storagetypes.Interface
	// IndexName is the name of the article index
	IndexName string
	// metrics holds processing metrics
	metrics *common.Metrics
}

// NewService creates a new article service.
func NewService(
	logger logger.Interface,
	defaultSelectors types.ArticleSelectors,
	storage storagetypes.Interface,
	indexName string,
) Interface {
	return &Service{
		Logger:           logger,
		SourceSelectors:  make(map[string]types.ArticleSelectors),
		DefaultSelectors: defaultSelectors,
		Storage:          storage,
		IndexName:        indexName,
		metrics:          &common.Metrics{},
	}
}

// AddSourceSelectors adds selectors for a specific source
func (s *Service) AddSourceSelectors(sourceName string, selectors types.ArticleSelectors) {
	s.SourceSelectors[sourceName] = selectors
}

// getSelectorsForURL returns the appropriate selectors for the given URL
func (s *Service) getSelectorsForURL(url string) types.ArticleSelectors {
	// Try to find matching source selectors
	for sourceName := range s.SourceSelectors {
		if strings.Contains(url, sourceName) {
			return s.SourceSelectors[sourceName]
		}
	}
	return s.DefaultSelectors
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

	// Get selectors for this URL
	selectors := s.getSelectorsForURL(e.Request.URL.String())

	// Extract metadata using the appropriate selectors
	article.Title = s.extractTitle(e, selectors)
	article.Description = s.extractDescription(e, selectors)
	article.PublishedDate = s.parsePublishedTime(e, selectors)
	article.Author = s.extractAuthor(e, selectors)
	article.Section = s.extractSection(e, selectors)
	article.CanonicalURL = s.extractCanonicalURL(e)

	return article
}

// ExtractContent extracts the main content from the HTML element
func (s *Service) ExtractContent(e *colly.HTMLElement, article *models.Article) {
	selectors := s.getSelectorsForURL(e.Request.URL.String())
	bodyEl := s.findArticleBody(e, selectors)
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

func (s *Service) extractTitle(e *colly.HTMLElement, selectors types.ArticleSelectors) string {
	// Try OpenGraph title first
	if title := e.ChildAttr(`meta[property="og:title"]`, "content"); title != "" {
		return title
	}

	// Try article title
	if title := e.ChildText(selectors.Title); title != "" {
		return title
	}

	// Fallback to page title
	return e.ChildText("title")
}

func (s *Service) extractDescription(e *colly.HTMLElement, selectors types.ArticleSelectors) string {
	// Try OpenGraph description
	if desc := e.ChildAttr(`meta[property="og:description"]`, "content"); desc != "" {
		return desc
	}

	// Try meta description
	return e.ChildAttr(selectors.Description, "content")
}

func (s *Service) extractPublishedTime(e *colly.HTMLElement, selectors types.ArticleSelectors) string {
	// Try article published publishedTime
	if publishedTime := e.ChildAttr(selectors.PublishedTime, "content"); publishedTime != "" {
		return publishedTime
	}

	// Try meta published publishedTime
	return e.ChildAttr(`meta[property="article:published_time"]`, "content")
}

func (s *Service) extractAuthor(e *colly.HTMLElement, selectors types.ArticleSelectors) string {
	// Try article author
	if author := e.ChildText(selectors.Byline + " " + selectors.Author); author != "" {
		return author
	}

	// Try meta author
	return e.ChildAttr(selectors.Author, "content")
}

func (s *Service) extractSection(e *colly.HTMLElement, selectors types.ArticleSelectors) string {
	return e.ChildText(selectors.Section)
}

func (s *Service) extractCanonicalURL(e *colly.HTMLElement) string {
	return e.ChildAttr("link[rel=canonical]", "href")
}

func (s *Service) findArticleBody(e *colly.HTMLElement, selectors types.ArticleSelectors) *goquery.Selection {
	bodySelector := selectors.Body
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

func (s *Service) parsePublishedTime(e *colly.HTMLElement, selectors types.ArticleSelectors) time.Time {
	timeStr := s.extractPublishedTime(e, selectors)
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

	// Log parsing error
	s.Logger.Debug("Failed to parse published time",
		"time", timeStr,
		"error", err)

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

// ExtractTags extracts tags from the HTML element and JSON-LD data
func (s *Service) ExtractTags(e *colly.HTMLElement, jsonLD JSONLDArticle) []string {
	selectors := s.getSelectorsForURL(e.Request.URL.String())
	var tags []string

	// Extract tags from meta keywords
	if keywords := e.ChildAttr(selectors.Keywords, "content"); keywords != "" {
		tags = append(tags, strings.Split(keywords, ",")...)
	}

	// Extract tags from JSON-LD keywords
	if len(jsonLD.Keywords) > 0 {
		tags = append(tags, jsonLD.Keywords...)
	}

	// Extract tags from article tags
	if articleTags := e.ChildText(selectors.Tags); articleTags != "" {
		tags = append(tags, strings.Split(articleTags, ",")...)
	}

	// Clean and deduplicate tags
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}

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

// ParsePublishedDate parses the published date from various sources
func (s *Service) ParsePublishedDate(e *colly.HTMLElement, jsonLD JSONLDArticle) time.Time {
	selectors := s.getSelectorsForURL(e.Request.URL.String())
	var dates []string

	// Try article published time
	if publishedTime := e.ChildAttr(selectors.PublishedTime, "content"); publishedTime != "" {
		dates = append(dates, publishedTime)
	}

	// Try meta published time
	if metaTime := e.ChildAttr(`meta[property="article:published_time"]`, "content"); metaTime != "" {
		dates = append(dates, metaTime)
	}

	// Try JSON-LD published time
	if jsonLD.DatePublished != "" {
		dates = append(dates, jsonLD.DatePublished)
	}

	return parseDate(dates, s.Logger)
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

// Process processes an article.
func (s *Service) Process(article *models.Article) error {
	if article == nil {
		return nil
	}

	// Store the article
	err := s.Storage.IndexDocument(context.Background(), s.IndexName, article.ID, article)
	if err != nil {
		s.Logger.Error("Failed to store article",
			"error", err,
			"articleID", article.ID,
			"url", article.Source)
		s.metrics.ErrorCount++
		return err
	}

	s.metrics.ProcessedCount++
	s.metrics.LastProcessedTime = time.Now()
	return nil
}

// ProcessJob processes a job and its items.
func (s *Service) ProcessJob(ctx context.Context, job *common.Job) {
	if job == nil {
		return
	}

	s.Logger.Info("Processing job",
		"jobID", job.ID,
		"url", job.URL)

	start := time.Now()
	defer func() {
		s.metrics.ProcessingDuration += time.Since(start)
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		s.Logger.Warn("Job processing cancelled",
			"jobID", job.ID,
			"error", ctx.Err())
		s.metrics.ErrorCount++
		return
	default:
		// Continue processing
	}
}

// ProcessHTML processes HTML content from a source.
func (s *Service) ProcessHTML(e *colly.HTMLElement) error {
	if e == nil {
		return nil
	}

	article := s.ExtractArticle(e)
	if article == nil {
		s.Logger.Debug("No article extracted",
			"url", e.Request.URL.String())
		return nil
	}

	return s.Process(article)
}

// GetMetrics returns the current processing metrics.
func (s *Service) GetMetrics() *common.Metrics {
	return s.metrics
}
