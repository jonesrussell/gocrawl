package sources_test

import (
	"os"
	"testing"

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

func TestLoadFromFile(t *testing.T) {
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

			_, err = tmpfile.WriteString(tt.content)
			require.NoError(t, err)

			err = tmpfile.Close()
			require.NoError(t, err)

			s, err := sources.LoadFromFile(tmpfile.Name())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, s.Sources, tt.want)
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		source  *sources.Config
		wantErr bool
	}{
		{
			name: "valid config",
			source: &sources.Config{
				Name:      "test",
				URL:       "http://example.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			source:  nil,
			wantErr: true,
		},
		{
			name: "missing name",
			source: &sources.Config{
				URL:       "http://example.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			source: &sources.Config{
				Name:      "test",
				RateLimit: "1s",
				MaxDepth:  2,
			},
			wantErr: true,
		},
		{
			name: "missing rate limit",
			source: &sources.Config{
				Name:     "test",
				URL:      "http://example.com",
				MaxDepth: 2,
			},
			wantErr: true,
		},
		{
			name: "invalid max depth",
			source: &sources.Config{
				Name:      "test",
				URL:       "http://example.com",
				RateLimit: "1s",
				MaxDepth:  0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sources.Sources{}
			err := s.Validate(tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
