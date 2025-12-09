// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the interface for processing articles.
// It handles the extraction and processing of article content from web pages.
type Interface interface {
	// Process handles the processing of article content.
	// It takes a colly.HTMLElement and processes the article found within it.
	Process(e *colly.HTMLElement) error

	// ProcessArticle processes an article and returns any errors
	ProcessArticle(ctx context.Context, article *domain.Article) error

	// Get retrieves an article by its ID
	Get(ctx context.Context, id string) (*domain.Article, error)

	// List retrieves a list of articles based on the provided query
	List(ctx context.Context, query map[string]any) ([]*domain.Article, error)

	// Delete removes an article by its ID
	Delete(ctx context.Context, id string) error

	// Update updates an existing article
	Update(ctx context.Context, article *domain.Article) error

	// Create creates a new article
	Create(ctx context.Context, article *domain.Article) error
}

// Ensure ContentService implements Interface
var _ Interface = (*ContentService)(nil)

// ContentService implements both Interface and ServiceInterface for article processing.
type ContentService struct {
	logger    logger.Interface
	storage   types.Interface
	indexName string
	sources   sources.Interface
	validator *ArticleValidator
}

// NewContentService creates a new article service.
func NewContentService(log logger.Interface, storage types.Interface, indexName string) *ContentService {
	return &ContentService{
		logger:    log,
		storage:   storage,
		indexName: indexName,
		validator: NewArticleValidator(log),
	}
}

// NewContentServiceWithSources creates a new article service with sources access.
func NewContentServiceWithSources(
	log logger.Interface,
	storage types.Interface,
	indexName string,
	sourcesManager sources.Interface,
) *ContentService {
	return &ContentService{
		logger:    log,
		storage:   storage,
		indexName: indexName,
		sources:   sourcesManager,
		validator: NewArticleValidator(log),
	}
}

// Process implements the Interface for HTML element processing.
func (s *ContentService) Process(e *colly.HTMLElement) error {
	if e == nil {
		return errors.New("HTML element is nil")
	}

	sourceURL := e.Request.URL.String()

	// Get source configuration and determine index name
	// Use local variable to avoid data race when Process() is called concurrently
	indexName := s.indexName
	var selectors configtypes.ArticleSelectors
	if s.sources != nil {
		// Try to find source by matching URL domain
		sourceConfig := s.findSourceByURL(sourceURL)
		if sourceConfig != nil {
			// Convert types.ArticleSelectors to configtypes.ArticleSelectors
			selectors = configtypes.ArticleSelectors{
				Container:     sourceConfig.Selectors.Article.Container,
				Title:         sourceConfig.Selectors.Article.Title,
				Body:          sourceConfig.Selectors.Article.Body,
				Intro:         sourceConfig.Selectors.Article.Intro,
				Link:          sourceConfig.Selectors.Article.Link,
				Image:         sourceConfig.Selectors.Article.Image,
				Byline:        sourceConfig.Selectors.Article.Byline,
				PublishedTime: sourceConfig.Selectors.Article.PublishedTime,
				TimeAgo:       sourceConfig.Selectors.Article.TimeAgo,
				JSONLD:        sourceConfig.Selectors.Article.JSONLD,
				Section:       sourceConfig.Selectors.Article.Section,
				Keywords:      sourceConfig.Selectors.Article.Keywords,
				Description:   sourceConfig.Selectors.Article.Description,
				OGTitle:       sourceConfig.Selectors.Article.OGTitle,
				OGDescription: sourceConfig.Selectors.Article.OGDescription,
				OGImage:       sourceConfig.Selectors.Article.OGImage,
				OGType:        sourceConfig.Selectors.Article.OGType,
				OGSiteName:    sourceConfig.Selectors.Article.OGSiteName,
				OgURL:         sourceConfig.Selectors.Article.OgURL,
				Canonical:     sourceConfig.Selectors.Article.Canonical,
				WordCount:     sourceConfig.Selectors.Article.WordCount,
				PublishDate:   sourceConfig.Selectors.Article.PublishDate,
				Category:      sourceConfig.Selectors.Article.Category,
				Tags:          sourceConfig.Selectors.Article.Tags,
				Author:        sourceConfig.Selectors.Article.Author,
				BylineName:    sourceConfig.Selectors.Article.BylineName,
				ArticleID:     sourceConfig.Selectors.Article.ArticleID,
				Exclude:       sourceConfig.Selectors.Article.Exclude,
			}
			// Use source's article index if available (local variable, no race condition)
			if sourceConfig.ArticleIndex != "" {
				indexName = sourceConfig.ArticleIndex
			}
		} else {
			s.logger.Debug("Source not found for URL, using default selectors",
				"url", sourceURL)
		}
	}

	// Extract article data using Colly methods
	articleData := extractArticle(e, selectors, sourceURL)

	// Clean category field
	categories := CleanCategory(articleData.Category)
	categoryStr := ""
	if len(categories) > 0 {
		categoryStr = categories[0] // Use first category as primary
	}

	// Calculate word count
	wordCount := CalculateWordCount(articleData.Body)

	// Convert to domain.Article
	article := &domain.Article{
		ID:            articleData.ID,
		Title:         articleData.Title,
		Body:          articleData.Body,
		Intro:         articleData.Intro,
		Author:        articleData.Author,
		BylineName:    articleData.BylineName,
		PublishedDate: articleData.PublishedDate,
		Source:        articleData.Source,
		Tags:          articleData.Tags,
		Keywords:      articleData.Keywords,
		Description:   articleData.Description,
		Section:       articleData.Section,
		Category:      categoryStr,
		OgTitle:       articleData.OgTitle,
		OgDescription: articleData.OgDescription,
		OgImage:       articleData.OgImage,
		OgURL:         articleData.OgURL,
		CanonicalURL:  articleData.CanonicalURL,
		WordCount:     wordCount,
		CreatedAt:     articleData.CreatedAt,
		UpdatedAt:     articleData.UpdatedAt,
	}

	// Validate article before indexing
	validationResult := s.validator.ValidateArticle(article)
	if !validationResult.IsValid {
		s.logger.Warn("Article validation failed, skipping index",
			"url", article.Source,
			"articleID", article.ID,
			"reason", validationResult.Reason,
			"title", article.Title)
		return fmt.Errorf("article validation failed: %s", validationResult.Reason)
	}

	// Process the article using the service interface with the determined index name
	return s.ProcessArticleWithIndex(context.Background(), article, indexName)
}

// findSourceByURL attempts to find a source configuration by matching the URL domain.
func (s *ContentService) findSourceByURL(pageURL string) *sources.Config {
	if s.sources == nil {
		return nil
	}

	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return nil
	}

	// Get all sources and try to match by domain
	sourceConfigs, err := s.sources.GetSources()
	if err != nil {
		return nil
	}

	for i := range sourceConfigs {
		source := &sourceConfigs[i]
		// Check if domain matches any allowed domain
		for _, allowedDomain := range source.AllowedDomains {
			if allowedDomain == hostname || allowedDomain == "*."+hostname {
				return source
			}
		}
		// Also check source URL
		if sourceParsedURL, parseErr := url.Parse(source.URL); parseErr == nil {
			if sourceParsedURL.Hostname() == hostname {
				return source
			}
		}
	}

	return nil
}

// ProcessArticle implements the ServiceInterface for article processing.
func (s *ContentService) ProcessArticle(ctx context.Context, article *domain.Article) error {
	return s.ProcessArticleWithIndex(ctx, article, s.indexName)
}

// ProcessArticleWithIndex processes an article and indexes it to the specified index.
// This method uses a local indexName parameter to avoid data races when called concurrently.
func (s *ContentService) ProcessArticleWithIndex(ctx context.Context, article *domain.Article, indexName string) error {
	if article == nil {
		return errors.New("article is nil")
	}

	if article.ID == "" {
		return errors.New("article ID is required")
	}

	if article.Source == "" {
		return errors.New("article source URL is required")
	}

	// Ensure word count is calculated if not set
	if article.WordCount == 0 {
		article.WordCount = CalculateWordCount(article.Body)
	}

	// Clean category if needed
	if article.Category != "" {
		categories := CleanCategory(article.Category)
		if len(categories) > 0 {
			article.Category = categories[0]
		} else {
			article.Category = ""
		}
	}

	// Validate article before indexing (if validator is set)
	if s.validator != nil {
		validationResult := s.validator.ValidateArticle(article)
		if !validationResult.IsValid {
			s.logger.Warn("Article validation failed, skipping index",
				"url", article.Source,
				"articleID", article.ID,
				"reason", validationResult.Reason,
				"title", article.Title)
			return fmt.Errorf("article validation failed: %s", validationResult.Reason)
		}
	}

	// Prepare article for indexing: clean empty fields, normalize arrays, prevent duplication
	article.PrepareForIndexing()

	// Index the article to Elasticsearch
	if err := s.storage.IndexDocument(ctx, indexName, article.ID, article); err != nil {
		s.logger.Error("Failed to index article",
			"error", err,
			"articleID", article.ID,
			"url", article.Source,
			"index", indexName)
		return fmt.Errorf("failed to index article: %w", err)
	}

	s.logger.Info("Article indexed successfully",
		"articleID", article.ID,
		"url", article.Source,
		"index", indexName,
		"title", article.Title,
		"wordCount", article.WordCount,
		"publishedDate", article.PublishedDate.Format(time.RFC3339),
		"category", article.Category)

	return nil
}

// Get implements the ServiceInterface.
func (s *ContentService) Get(ctx context.Context, id string) (*domain.Article, error) {
	// Implementation
	return nil, errors.New("not implemented")
}

// List returns a list of articles matching the query
func (s *ContentService) List(ctx context.Context, query map[string]any) ([]*domain.Article, error) {
	// TODO: Implement article listing
	return nil, errors.New("not implemented")
}

// Delete implements the ServiceInterface.
func (s *ContentService) Delete(ctx context.Context, id string) error {
	// Implementation
	return nil
}

// Update implements the ServiceInterface.
func (s *ContentService) Update(ctx context.Context, article *domain.Article) error {
	// Implementation
	return nil
}

// Create implements the ServiceInterface.
func (s *ContentService) Create(ctx context.Context, article *domain.Article) error {
	// Implementation
	return nil
}
