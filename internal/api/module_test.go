package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSearchService struct {
	mock.Mock
}

func (m *mockSearchService) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	args := m.Called(ctx, query, size)
	return args.Get(0).([]*models.Article), args.Error(1)
}

func TestStartHTTPServer(t *testing.T) {
	// Create mock logger
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "StartHTTPServer function called").Return()

	// Create mock search service
	mockSearch := new(mockSearchService)
	mockArticles := []*models.Article{
		{ID: "1", Title: "Test Article"},
	}
	mockSearch.On("SearchArticles", mock.Anything, "test query", 10).Return(mockArticles, nil)

	// Start the server
	server, err := api.StartHTTPServer(mockLogger, mockSearch)
	require.NoError(t, err)
	assert.NotNil(t, server)

	// Create test request
	req := api.SearchRequest{
		Query: "test query",
		Size:  10,
	}
	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	// Serve the request
	server.Handler.ServeHTTP(recorder, request)

	// Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []*models.Article
	err = json.NewDecoder(recorder.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, mockArticles, response)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}
