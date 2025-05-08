package models

import (
	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/mock"
)

// MockPageProcessor is a mock implementation of PageProcessor
type MockPageProcessor struct {
	mock.Mock
}

// Process implements PageProcessor
func (m *MockPageProcessor) Process(e *colly.HTMLElement) {
	m.Called(e)
}
