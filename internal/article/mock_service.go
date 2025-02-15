package article

import (
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the article.Interface
type MockService struct {
	mock.Mock
}

// ExtractArticle implements Service.ExtractArticle
func (m *MockService) ExtractArticle(e *colly.HTMLElement) *models.Article {
	args := m.Called(e)
	return args.Get(0).(*models.Article)
}

// ExtractTags mocks the ExtractTags method
func (m *MockService) ExtractTags(e *colly.HTMLElement, jsonLD JSONLDArticle) []string {
	args := m.Called(e, jsonLD)
	return args.Get(0).([]string)
}

// CleanAuthor mocks the CleanAuthor method
func (m *MockService) CleanAuthor(author string) string {
	args := m.Called(author)
	return args.String(0)
}

// ParsePublishedDate mocks the ParsePublishedDate method
func (m *MockService) ParsePublishedDate(e *colly.HTMLElement, jsonLD JSONLDArticle) time.Time {
	args := m.Called(e, jsonLD)
	return args.Get(0).(time.Time)
}
