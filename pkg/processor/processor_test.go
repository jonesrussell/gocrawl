package processor_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/pkg/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name    string
		params  processor.Params
		wantErr bool
	}{
		{
			name: "valid params",
			params: processor.Params{
				Logger: logger.NewNoOp(),
			},
			wantErr: false,
		},
		{
			name:    "missing logger",
			params:  processor.Params{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := processor.NewProcessor(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, p)
		})
	}
}

func TestProcessor_Process(t *testing.T) {
	params := processor.Params{
		Logger: logger.NewNoOp(),
	}
	p, err := processor.NewProcessor(params)
	require.NoError(t, err)
	require.NotNil(t, p)

	tests := []struct {
		name    string
		source  string
		content []byte
		wantErr bool
	}{
		{
			name:   "valid content",
			source: "test",
			content: []byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<link rel="canonical" href="http://example.com/article">
					<meta property="og:url" content="http://example.com/article">
					<meta name="description" content="Test article">
				</head>
				<body>
					<h1>Test Article</h1>
					<article>This is the article body.</article>
					<div class="author">John Doe</div>
					<time>2024-03-20T10:00:00Z</time>
					<div class="categories">
						<span>Technology</span>
						<span>Programming</span>
					</div>
					<div class="tags">
						<span>go</span>
						<span>testing</span>
					</div>
				</body>
				</html>
			`),
			wantErr: false,
		},
		{
			name:    "empty source",
			source:  "",
			content: []byte("<h1>Test</h1>"),
			wantErr: true,
		},
		{
			name:    "empty content",
			source:  "test",
			content: []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processErr := p.Process(t.Context(), tt.source, tt.content)
			if tt.wantErr {
				require.Error(t, processErr)
				return
			}
			require.NoError(t, processErr)
		})
	}
}

func TestProcessor_Validate(t *testing.T) {
	params := processor.Params{
		Logger: logger.NewNoOp(),
	}
	p, err := processor.NewProcessor(params)
	require.NoError(t, err)
	require.NotNil(t, p)

	tests := []struct {
		name    string
		source  string
		content []byte
		wantErr bool
	}{
		{
			name:    "valid content",
			source:  "test",
			content: []byte("<h1>Test</h1>"),
			wantErr: false,
		},
		{
			name:    "empty source",
			source:  "",
			content: []byte("<h1>Test</h1>"),
			wantErr: true,
		},
		{
			name:    "empty content",
			source:  "test",
			content: []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateErr := p.Validate(tt.source, tt.content)
			if tt.wantErr {
				require.Error(t, validateErr)
				return
			}
			require.NoError(t, validateErr)
		})
	}
}

func TestProcessor_GetMetrics(t *testing.T) {
	params := processor.Params{
		Logger: logger.NewNoOp(),
	}
	p, err := processor.NewProcessor(params)
	require.NoError(t, err)
	require.NotNil(t, p)

	// Process some content to update metrics
	processErr := p.Process(t.Context(), "test", []byte(`
		<!DOCTYPE html>
		<html>
		<body>
			<h1>Test Article</h1>
			<article>This is the article body.</article>
		</body>
		</html>
	`))
	require.NoError(t, processErr)

	metrics := p.GetMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.ProcessedCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.NotZero(t, metrics.ProcessingDuration)
	assert.NotZero(t, metrics.LastProcessedTime)
}
