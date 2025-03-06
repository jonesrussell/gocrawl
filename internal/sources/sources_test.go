package sources_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/require"
)

// MockLogger is a mock implementation of logger.Interface
type MockLogger struct {
	startCalled        bool
	startMsg           string
	startKeysAndValues []interface{}
	errorCalled        bool
	errorMsg           string
	errorKeysAndValues []interface{}
	debugCalled        bool
	debugMsg           string
	debugKeysAndValues []interface{}
	warnCalled         bool
	warnMsg            string
	warnKeysAndValues  []interface{}
	errorfCalled       bool
	errorfFormat       string
	errorfArgs         []interface{}
	fatalCalled        bool
	fatalMsg           string
	fatalKeysAndValues []interface{}
	printfCalled       bool
	printfFormat       string
	printfArgs         []interface{}
	syncCalled         bool
	syncErr            error
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.startCalled = true
	m.startMsg = msg
	m.startKeysAndValues = keysAndValues
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.errorCalled = true
	m.errorMsg = msg
	m.errorKeysAndValues = keysAndValues
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.debugCalled = true
	m.debugMsg = msg
	m.debugKeysAndValues = keysAndValues
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.warnCalled = true
	m.warnMsg = msg
	m.warnKeysAndValues = keysAndValues
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.errorfCalled = true
	m.errorfFormat = format
	m.errorfArgs = args
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.fatalCalled = true
	m.fatalMsg = msg
	m.fatalKeysAndValues = keysAndValues
}

func (m *MockLogger) Printf(format string, args ...interface{}) {
	m.printfCalled = true
	m.printfFormat = format
	m.printfArgs = args
}

func (m *MockLogger) Sync() error {
	m.syncCalled = true
	return m.syncErr
}

type mockCrawler struct {
	startCalled bool
	startURL    string
	startErr    error
}

func (m *mockCrawler) Start(ctx context.Context, url string) error {
	m.startCalled = true
	m.startURL = url
	return m.startErr
}

func (m *mockCrawler) Stop() {
	m.startCalled = false
}

type mockIndexManager struct {
	ensureIndexCalled bool
	ensureIndexName   string
	ensureIndexErr    error
}

func (m *mockIndexManager) EnsureIndex(ctx context.Context, name string) error {
	m.ensureIndexCalled = true
	m.ensureIndexName = name
	return m.ensureIndexErr
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name: "valid sources",
			content: `sources:
  - name: test1
    url: http://example1.com
  - name: test2
    url: http://example2.com`,
			want:    2,
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			content: "invalid: [yaml: content",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp(t.TempDir(), "sources*.yml")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.Write([]byte(tt.content))
			require.NoError(t, err)

			err = tmpfile.Close()
			require.NoError(t, err)

			s, err := sources.Load(tmpfile.Name())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, len(s.Sources))
		})
	}
}

func TestFindByName(t *testing.T) {
	sources := &sources.Sources{
		Sources: []sources.Config{
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
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantName, source.Name)
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		mockErr   error
		wantError bool
	}{
		{
			name:      "successful start",
			source:    "test",
			mockErr:   nil,
			wantError: false,
		},
		{
			name:      "crawler error",
			source:    "test",
			mockErr:   errors.New("crawler error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			crawler := &mockCrawler{startErr: tt.mockErr}
			indexMgr := &mockIndexManager{}

			s := &sources.Sources{
				Sources: []sources.Config{
					{
						Name:         "test",
						URL:          "http://example.com",
						Index:        "test_index",
						ArticleIndex: "articles",
					},
				},
				Logger:   &logger.MockLogger{},
				Crawler:  crawler,
				IndexMgr: indexMgr,
			}

			err := s.Start(ctx, tt.source)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.True(t, crawler.startCalled)
			require.Equal(t, "http://example.com", crawler.startURL)
			require.True(t, indexMgr.ensureIndexCalled)
			require.Equal(t, "test_index", indexMgr.ensureIndexName)
		})
	}
}

func TestStop(t *testing.T) {
	crawler := &mockCrawler{}
	s := &sources.Sources{
		Crawler: crawler,
	}

	crawler.startCalled = true
	s.Stop()
	require.False(t, crawler.startCalled)

	// Test with nil crawler
	s = &sources.Sources{}
	s.Stop() // Should not panic
}
