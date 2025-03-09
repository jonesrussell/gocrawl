package models

import (
	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/mock"
)

// MockContentProcessor is a mock implementation of ContentProcessor
type MockContentProcessor struct {
	mock.Mock
}

// Process implements ContentProcessor
func (m *MockContentProcessor) Process(e *colly.HTMLElement) {
	m.Called(e)
}

// NewMockContentProcessor creates a new instance of MockContentProcessor
func NewMockContentProcessor() *MockContentProcessor {
	return &MockContentProcessor{}
}
