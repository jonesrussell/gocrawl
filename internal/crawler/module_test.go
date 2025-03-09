package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestProvideCrawler(t *testing.T) {
	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name: "missing logger",
			params: Params{
				IndexManager: api.NewMockIndexManager(),
			},
			wantErr: true,
		},
		{
			name: "missing index manager",
			params: Params{
				Logger: logger.NewMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: Params{
				Logger:       logger.NewMockLogger(),
				IndexManager: api.NewMockIndexManager(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := provideEventBus()
			result, err := provideCrawler(tt.params, bus)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result.Crawler)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result.Crawler)
		})
	}
}

func TestProvideCollyDebugger(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	debugger := provideCollyDebugger(mockLogger)
	assert.NotNil(t, debugger)
}
