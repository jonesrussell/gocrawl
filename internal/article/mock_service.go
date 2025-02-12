package article

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

// ExtractArticle implements Service.ExtractArticle
func (m *MockService) ExtractArticle(e *colly.HTMLElement) *models.Article {
	args := m.Called(e)
	if article, ok := args.Get(0).(*models.Article); ok {
		return article
	}
	return nil
}
