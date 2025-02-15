package article_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	mockLogger := logger.NewMockLogger() // Assuming you have a mock logger
	svc := article.NewService(mockLogger)

	assert.NotNil(t, svc)
}
