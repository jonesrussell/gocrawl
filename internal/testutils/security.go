// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockSecurityMiddleware implements SecurityMiddlewareInterface for testing
type MockSecurityMiddleware struct {
	mock.Mock
}

func (m *MockSecurityMiddleware) Cleanup(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockSecurityMiddleware) WaitCleanup() {
	m.Called()
}

func (m *MockSecurityMiddleware) Middleware() gin.HandlerFunc {
	args := m.Called()
	if fn := args.Get(0); fn != nil {
		return fn.(gin.HandlerFunc)
	}
	return func(c *gin.Context) { c.Next() }
}
