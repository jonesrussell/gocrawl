package sources

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of logger.Interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Printf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// MockCrawler is a mock implementation of the Crawler interface
type MockCrawler struct {
	mock.Mock
}

func NewMockCrawler() *MockCrawler {
	return &MockCrawler{}
}

func (m *MockCrawler) Start(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockCrawler) Stop() {
	m.Called()
}

// MockIndexManager is a mock implementation of IndexManager
type MockIndexManager struct {
	mock.Mock
}

func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}

func (m *MockIndexManager) EnsureIndex(ctx context.Context, indexName string) error {
	args := m.Called(ctx, indexName)
	return args.Error(0)
}

func TestLoad(t *testing.T) {
	// Create a temporary test file
	content := `sources:
  - name: test
    url: http://example.com
    index: test_index
    rate_limit: 1s
    max_depth: 2
    time:
      - "09:00"
      - "15:00"`

	tmpfile, err := os.CreateTemp("", "sources*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "valid file",
			file:    tmpfile.Name(),
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    "nonexistent.yml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sources, err := Load(tt.file)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, sources)
			if !tt.wantErr {
				assert.Len(t, sources.Sources, 1)
				assert.Equal(t, "test", sources.Sources[0].Name)
				assert.Equal(t, "http://example.com", sources.Sources[0].URL)
				assert.Equal(t, "test_index", sources.Sources[0].Index)
				assert.Equal(t, "1s", sources.Sources[0].RateLimit)
				assert.Equal(t, 2, sources.Sources[0].MaxDepth)
				assert.Equal(t, []string{"09:00", "15:00"}, sources.Sources[0].Time)
			}
		})
	}
}

func TestFindByName(t *testing.T) {
	sources := &Sources{
		Sources: []Config{
			{
				Name: "test1",
				URL:  "http://example1.com",
			},
			{
				Name: "test2",
				URL:  "http://example2.com",
			},
		},
	}

	tests := []struct {
		name     string
		source   string
		wantName string
		wantErr  bool
	}{
		{
			name:     "existing source",
			source:   "test1",
			wantName: "test1",
			wantErr:  false,
		},
		{
			name:    "non-existent source",
			source:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := sources.FindByName(tt.source)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantName, source.Name)
		})
	}
}

func TestStart(t *testing.T) {
	mockLogger := new(MockLogger)
	mockCrawler := NewMockCrawler()
	mockIndexMgr := NewMockIndexManager()

	sources := &Sources{
		Sources: []Config{
			{
				Name:         "test1",
				URL:          "http://example1.com",
				ArticleIndex: "test1_articles",
				Index:        "test1_content",
			},
		},
		Logger:   mockLogger,
		Crawler:  mockCrawler,
		IndexMgr: mockIndexMgr,
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		source    string
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "successful crawl",
			source: "test1",
			setupMock: func() {
				mockLogger.On("Info", "Starting crawl", []interface{}{"source", "test1"}).Once()
				mockIndexMgr.On("EnsureIndex", ctx, "test1_articles").Return(nil).Once()
				mockIndexMgr.On("EnsureIndex", ctx, "test1_content").Return(nil).Once()
				mockCrawler.On("Start", ctx, "http://example1.com").Return(nil).Once()
				mockLogger.On("Info", "Finished crawl", []interface{}{"source", "test1"}).Once()
			},
			wantErr: false,
		},
		{
			name:   "crawler error",
			source: "test1",
			setupMock: func() {
				mockLogger.On("Info", "Starting crawl", []interface{}{"source", "test1"}).Once()
				mockIndexMgr.On("EnsureIndex", ctx, "test1_articles").Return(nil).Once()
				mockIndexMgr.On("EnsureIndex", ctx, "test1_content").Return(nil).Once()
				mockCrawler.On("Start", ctx, "http://example1.com").Return(assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name:      "non-existent source",
			source:    "nonexistent",
			setupMock: func() {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := sources.Start(ctx, tt.source)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockLogger.AssertExpectations(t)
			mockCrawler.AssertExpectations(t)
			mockIndexMgr.AssertExpectations(t)
		})
	}
}

func TestStop(t *testing.T) {
	mockCrawler := NewMockCrawler()
	sources := &Sources{
		Crawler: mockCrawler,
	}

	mockCrawler.On("Stop").Once()
	sources.Stop()
	mockCrawler.AssertExpectations(t)

	// Test with nil crawler
	sources = &Sources{}
	sources.Stop() // Should not panic
}
