// Package articles provides functionality for processing and managing article content.
package articles

import (
	"context"
	"errors"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentService implements both Interface and ServiceInterface for article processing.
type ContentService struct {
	logger    logger.Interface
	storage   types.Interface
	indexName string
}

// NewContentService creates a new article service.
func NewContentService(logger logger.Interface, storage types.Interface, indexName string) *ContentService {
	return &ContentService{
		logger:    logger,
		storage:   storage,
		indexName: indexName,
	}
}

// Process implements the Interface for HTML element processing.
func (s *ContentService) Process(e *colly.HTMLElement) error {
	// Extract article data from the HTML element
	article := &models.Article{
		Source:    e.Request.URL.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Process the article using the service interface
	return s.ProcessArticle(context.Background(), article)
}

// ProcessArticle implements the ServiceInterface for article processing.
func (s *ContentService) ProcessArticle(ctx context.Context, article *models.Article) error {
	// Implementation
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
