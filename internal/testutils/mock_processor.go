// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/stretchr/testify/mock"
)

// MockProcessor implements collector.Processor for testing
type MockProcessor struct {
	mock.Mock
	ProcessCalls int
}

// Process implements collector.Processor
func (m *MockProcessor) Process(e *colly.HTMLElement) error {
	m.ProcessCalls++
	args := m.Called(e)
	return args.Error(0)
}

// Ensure MockProcessor implements collector.Processor
var _ collector.Processor = (*MockProcessor)(nil)

// NewMockProcessor creates a new mock processor instance.
func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}
