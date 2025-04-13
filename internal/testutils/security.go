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
		if handler, ok := fn.(gin.HandlerFunc); ok {
			return handler
		}
	}
	return func(c *gin.Context) { c.Next() }
}

// GetMiddleware returns the middleware function
func (m *MockSecurityMiddleware) GetMiddleware() gin.HandlerFunc {
	args := m.Called()
	if len(args) == 0 {
		return nil
	}
	if fn, ok := args.Get(0).(gin.HandlerFunc); ok {
		return fn
	}
	return nil
}
