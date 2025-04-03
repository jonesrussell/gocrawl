package testutils

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
)

// MockProcessor implements common.Processor for testing
type MockProcessor struct {
	ProcessFunc     func(ctx context.Context, data interface{}) error
	CanProcessFunc  func(data interface{}) bool
	ContentTypeFunc func() common.ContentType
	GetMetricsFunc  func() *common.Metrics
	ProcessHTMLFunc func(ctx context.Context, html *colly.HTMLElement) error
	ProcessJobFunc  func(ctx context.Context, job *common.Job)
	StartFunc       func(ctx context.Context) error
	StopFunc        func(ctx context.Context) error
}

func (p *MockProcessor) Process(ctx context.Context, data interface{}) error {
	if p.ProcessFunc != nil {
		return p.ProcessFunc(ctx, data)
	}
	return nil
}

func (p *MockProcessor) CanProcess(data interface{}) bool {
	if p.CanProcessFunc != nil {
		return p.CanProcessFunc(data)
	}
	return true
}

func (p *MockProcessor) ContentType() common.ContentType {
	if p.ContentTypeFunc != nil {
		return p.ContentTypeFunc()
	}
	return common.ContentType("text/html")
}

func (p *MockProcessor) GetMetrics() *common.Metrics {
	if p.GetMetricsFunc != nil {
		return p.GetMetricsFunc()
	}
	return &common.Metrics{}
}

func (p *MockProcessor) ProcessHTML(ctx context.Context, html *colly.HTMLElement) error {
	if p.ProcessHTMLFunc != nil {
		return p.ProcessHTMLFunc(ctx, html)
	}
	return nil
}

func (p *MockProcessor) ProcessJob(ctx context.Context, job *common.Job) {
	if p.ProcessJobFunc != nil {
		p.ProcessJobFunc(ctx, job)
	}
}

func (p *MockProcessor) Start(ctx context.Context) error {
	if p.StartFunc != nil {
		return p.StartFunc(ctx)
	}
	return nil
}

func (p *MockProcessor) Stop(ctx context.Context) error {
	if p.StopFunc != nil {
		return p.StopFunc(ctx)
	}
	return nil
}

// NewMockProcessor creates a new mock processor
func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}

// ProvideMockProcessors provides mock processors for testing
func ProvideMockProcessors() (common.Processor, common.Processor) {
	return NewMockProcessor(), NewMockProcessor()
}
