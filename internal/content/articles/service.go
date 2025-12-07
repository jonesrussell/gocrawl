// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentService implements both Interface and ServiceInterface for article processing.
type ContentService struct {
	logger    logger.Interface
	storage   types.Interface
	indexName string
	sources   sources.Interface
}

// NewContentService creates a new article service.
func NewContentService(logger logger.Interface, storage types.Interface, indexName string) *ContentService {
	return &ContentService{
		logger:    logger,
		storage:   storage,
		indexName: indexName,
	}
}

// NewContentServiceWithSources creates a new article service with sources access.
func NewContentServiceWithSources(logger logger.Interface, storage types.Interface, indexName string, sources sources.Interface) *ContentService {
	return &ContentService{
		logger:    logger,
		storage:   storage,
		indexName: indexName,
		sources:   sources,
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
			// Convert sourceutils.ArticleSelectors to configtypes.ArticleSelectors
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

	// Convert to models.Article
	article := &models.Article{
		ID:            articleData.ID,
		Title:         articleData.Title,
		Body:          articleData.Body,
		Intro:         articleData.Intro,
		Author:        articleData.Author,
		BylineName:    articleData.BylineName,
		PublishedDate: articleData.PublishedDate,
		Source:        articleData.Source,
		Tags:          articleData.Tags,
		Description:   articleData.Description,
		Section:       articleData.Section,
		Category:      articleData.Category,
		OgTitle:       articleData.OgTitle,
		OgDescription: articleData.OgDescription,
		OgImage:       articleData.OgImage,
		OgURL:         articleData.OgURL,
		CanonicalURL:  articleData.CanonicalURL,
		CreatedAt:     articleData.CreatedAt,
		UpdatedAt:     articleData.UpdatedAt,
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

	domain := parsedURL.Hostname()
	if domain == "" {
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
			if allowedDomain == domain || allowedDomain == "*."+domain {
				return source
			}
		}
		// Also check source URL
		if sourceURL, err := url.Parse(source.URL); err == nil {
			if sourceURL.Hostname() == domain {
				return source
			}
		}
	}

	return nil
}

// ProcessArticle implements the ServiceInterface for article processing.
func (s *ContentService) ProcessArticle(ctx context.Context, article *models.Article) error {
	return s.ProcessArticleWithIndex(ctx, article, s.indexName)
}

// ProcessArticleWithIndex processes an article and indexes it to the specified index.
// This method uses a local indexName parameter to avoid data races when called concurrently.
func (s *ContentService) ProcessArticleWithIndex(ctx context.Context, article *models.Article, indexName string) error {
	if article == nil {
		return errors.New("article is nil")
	}

	if article.ID == "" {
		return errors.New("article ID is required")
	}

	if article.Source == "" {
		return errors.New("article source URL is required")
	}

	// Index the article to Elasticsearch
	if err := s.storage.IndexDocument(ctx, indexName, article.ID, article); err != nil {
		s.logger.Error("Failed to index article",
			"error", err,
			"articleID", article.ID,
			"url", article.Source,
			"index", indexName)
		return fmt.Errorf("failed to index article: %w", err)
	}

	s.logger.Debug("Article indexed successfully",
		"articleID", article.ID,
		"url", article.Source,
		"index", indexName,
		"title", article.Title)

	return nil
}

// Get implements the ServiceInterface.
func (s *ContentService) Get(ctx context.Context, id string) (*models.Article, error) {
	// Implementation
	return nil, errors.New("not implemented")
}

// List returns a list of articles matching the query
func (s *ContentService) List(ctx context.Context, query map[string]any) ([]*models.Article, error) {
	// TODO: Implement article listing
	return nil, errors.New("not implemented")
}

// Delete implements the ServiceInterface.
func (s *ContentService) Delete(ctx context.Context, id string) error {
	// Implementation
	return nil
}

// Update implements the ServiceInterface.
func (s *ContentService) Update(ctx context.Context, article *models.Article) error {
	// Implementation
	return nil
}

// Create implements the ServiceInterface.
func (s *ContentService) Create(ctx context.Context, article *models.Article) error {
	// Implementation
	return nil
}
