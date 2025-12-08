package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/constants"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Storage implements the ArticleStorage interface using the underlying storage implementation.
type Storage struct {
	logger    logger.Interface
	storage   storagetypes.Interface
	indexName string
}

// NewStorage creates a new Storage instance.
func NewStorage(
	log logger.Interface,
	storageInterface storagetypes.Interface,
	indexName string,
) *Storage {
	return &Storage{
		logger:    log,
		storage:   storageInterface,
		indexName: indexName,
	}
}

// SaveArticle saves an article to storage.
func (s *Storage) SaveArticle(ctx context.Context, article *domain.Article) error {
	if article == nil {
		return errors.New("article is nil")
	}

	if err := s.storage.IndexDocument(ctx, s.indexName, article.ID, article); err != nil {
		s.logger.Error("Failed to save article",
			"error", err,
			"articleID", article.ID,
			"url", article.Source)
		return fmt.Errorf("failed to save article: %w", err)
	}

	s.logger.Debug("Saved article",
		"articleID", article.ID,
		"url", article.Source)
	return nil
}

// GetArticle retrieves an article from storage.
func (s *Storage) GetArticle(ctx context.Context, id string) (*domain.Article, error) {
	if id == "" {
		return nil, errors.New("article ID is empty")
	}

	article := &domain.Article{}
	if err := s.storage.GetDocument(ctx, s.indexName, id, article); err != nil {
		s.logger.Error("Failed to get article",
			"error", err,
			"articleID", id)
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	s.logger.Debug("Retrieved article",
		"articleID", id)
	return article, nil
}

// ListArticles lists articles matching the query.
func (s *Storage) ListArticles(ctx context.Context, query string) ([]*domain.Article, error) {
	// Create a search query
	searchQuery := s.createSearchQuery(query)

	// Execute the search
	results, err := s.storage.Search(ctx, s.indexName, searchQuery)
	if err != nil {
		s.logger.Error("Failed to list articles",
			"error", err,
			"query", query)
		return nil, fmt.Errorf("failed to list articles: %w", err)
	}

	// Convert results to articles
	articles := s.convertResultsToArticles(results)

	s.logger.Debug("Listed articles",
		"query", query,
		"count", len(articles))
	return articles, nil
}

// createSearchQuery creates a search query for articles.
func (s *Storage) createSearchQuery(query string) map[string]any {
	return map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title^2", "body", "description"},
			},
		},
		"size": constants.DefaultBufferSize,
	}
}

// convertResultsToArticles converts search results to articles.
func (s *Storage) convertResultsToArticles(results []any) []*domain.Article {
	articles := make([]*domain.Article, 0, len(results))
	for _, result := range results {
		article, err := s.convertResultToArticle(result)
		if err != nil {
			continue
		}
		articles = append(articles, article)
	}
	return articles
}

// convertResultToArticle converts a single result to an article.
func (s *Storage) convertResultToArticle(result any) (*domain.Article, error) {
	if article, isArticle := result.(*domain.Article); isArticle {
		return article, nil
	}

	if m, isMap := result.(map[string]any); isMap {
		newArticle := &domain.Article{}
		data, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		if unmarshalErr := json.Unmarshal(data, newArticle); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal article: %w", unmarshalErr)
		}
		return newArticle, nil
	}

	return nil, fmt.Errorf("unsupported result type: %T", result)
}

// Store stores the result in the appropriate storage.
func (s *Storage) Store(ctx context.Context, result any) error {
	if result == nil {
		return errors.New("result cannot be nil")
	}

	// Handle article storage
	if article, isArticle := result.(*domain.Article); isArticle {
		if article == nil {
			return errors.New("article cannot be nil")
		}

		// Store article
		if err := s.SaveArticle(ctx, article); err != nil {
			return fmt.Errorf("failed to save article: %w", err)
		}

		return nil
	}

	return fmt.Errorf("unsupported result type: %T", result)
}

// Ensure Storage implements ArticleStorage interface
var _ ArticleStorage = (*Storage)(nil)
