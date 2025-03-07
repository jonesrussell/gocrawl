package storage_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchService_SearchArticles(t *testing.T) {
	// Create a mock storage instance
	mockStorage := storage.NewMockStorage()
	searchService := storage.NewSearchService(mockStorage)

	// Define test cases
	tests := []struct {
		name             string
		query            string
		size             int
		expectedArticles []*models.Article
		expectedError    error
	}{
		{
			name:             "Successful search",
			query:            "test",
			size:             10,
			expectedArticles: []*models.Article{{ID: "1", Title: "Test Article 1"}},
			expectedError:    nil,
		},
		{
			name:             "Search returns error",
			query:            "error",
			size:             5,
			expectedArticles: nil,
			expectedError:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock expectations
			mockStorage.ExpectedCalls = nil

			// Set up expectations with a more flexible context matcher
			mockStorage.On("SearchArticles", mock.Anything, tt.query, tt.size).Return(tt.expectedArticles, tt.expectedError)
			mockStorage.On("TestConnection", mock.Anything).Return(nil)
			mockStorage.On("IndexExists", mock.Anything, "articles").Return(true, nil)
			mockStorage.On("Close").Return(nil).Maybe()

			// Call the method under test
			articles, err := searchService.SearchArticles(t.Context(), tt.query, tt.size)

			// Assert results
			assert.Equal(t, tt.expectedArticles, articles)
			assert.Equal(t, tt.expectedError, err)

			// Assert that the expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}
