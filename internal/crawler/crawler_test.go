package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestCrawler_Stop(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	c := &crawler.Crawler{
		Logger: mockLogger,
	}

	ctx := context.Background()
	err := c.Stop(ctx)
	require.NoError(t, err)
}

func TestCrawler_SetCollector(t *testing.T) {
	c := &crawler.Crawler{}
	collector := &colly.Collector{}
	c.SetCollector(collector)
	// We can't test the private field directly
}

func TestCrawler_Start(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		setupMock func(*colly.Collector)
		wantErr   bool
	}{
		{
			name:    "empty base URL",
			baseURL: "",
			setupMock: func(_ *colly.Collector) {
				// No setup needed for empty URL test
			},
			wantErr: true,
		},
		{
			name:    "valid URL",
			baseURL: "http://example.com",
			setupMock: func(c *colly.Collector) {
				// Setup any collector mocks if needed
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := logger.NewMockLogger()
			c := &crawler.Crawler{
				Logger: mockLogger,
			}
			collector := colly.NewCollector()
			tt.setupMock(collector)
			c.SetCollector(collector)

			ctx := context.Background()
			err := c.Start(ctx, tt.baseURL)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCrawler_Start_ContextCancellation(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	c := &crawler.Crawler{
		Logger: mockLogger,
	}
	collector := colly.NewCollector()
	c.SetCollector(collector)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := c.Start(ctx, "http://example.com")
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}

func TestCrawler_Subscribe(t *testing.T) {
	c := &crawler.Crawler{}
	handler := func(ctx context.Context, content *events.Content) error {
		return nil
	}
	c.Subscribe(handler)
	// We can't test the private bus field directly
}

func TestCrawler_SetRateLimit(t *testing.T) {
	c := &crawler.Crawler{}
	tests := []struct {
		name    string
		limit   string
		wantErr bool
	}{
		{
			name:    "valid duration",
			limit:   "1s",
			wantErr: false,
		},
		{
			name:    "invalid duration",
			limit:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.SetRateLimit(tt.limit)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCrawler_SetMaxDepth(t *testing.T) {
	c := &crawler.Crawler{}
	c.SetMaxDepth(5)
	// We can't test the private collector field directly
}

func TestCrawler_GetIndexManager(t *testing.T) {
	c := &crawler.Crawler{}
	// We can't set the private field directly, but we can test the getter
	result := c.GetIndexManager()
	require.Nil(t, result) // Should be nil since we didn't set it
}

type mockIndexManager struct {
	mock.Mock
}

func (m *mockIndexManager) EnsureIndex(ctx context.Context, name string, mapping interface{}) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

func (m *mockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockIndexManager) UpdateMapping(ctx context.Context, name string, mapping interface{}) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}
