package crawler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

func TestProvideCrawler(t *testing.T) {
	mockProcessor := models.NewMockContentProcessor()
	mockProcessor.On("Process", mock.Anything).Return()

	tests := []struct {
		name    string
		params  crawler.Params
		wantErr bool
	}{
		{
			name: "missing logger",
			params: crawler.Params{
				Config:           &config.Config{},
				Storage:          &storage.MockStorage{},
				IndexManager:     api.NewMockIndexManager(),
				ContentProcessor: []models.ContentProcessor{mockProcessor},
			},
			wantErr: true,
		},
		{
			name: "missing config",
			params: crawler.Params{
				Logger:           logger.NewMockLogger(),
				Storage:          &storage.MockStorage{},
				IndexManager:     api.NewMockIndexManager(),
				ContentProcessor: []models.ContentProcessor{mockProcessor},
			},
			wantErr: true,
		},
		{
			name: "missing storage",
			params: crawler.Params{
				Logger:           logger.NewMockLogger(),
				Config:           &config.Config{},
				IndexManager:     api.NewMockIndexManager(),
				ContentProcessor: []models.ContentProcessor{mockProcessor},
			},
			wantErr: true,
		},
		{
			name: "missing index manager",
			params: crawler.Params{
				Logger:           logger.NewMockLogger(),
				Config:           &config.Config{},
				Storage:          &storage.MockStorage{},
				ContentProcessor: []models.ContentProcessor{mockProcessor},
			},
			wantErr: true,
		},
		{
			name: "missing content processor",
			params: crawler.Params{
				Logger:       logger.NewMockLogger(),
				Config:       &config.Config{},
				Storage:      &storage.MockStorage{},
				IndexManager: api.NewMockIndexManager(),
			},
			wantErr: true,
		},
		{
			name: "successful creation",
			params: crawler.Params{
				Logger:           logger.NewMockLogger(),
				Config:           &config.Config{},
				Storage:          &storage.MockStorage{},
				IndexManager:     api.NewMockIndexManager(),
				ContentProcessor: []models.ContentProcessor{mockProcessor},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mockLogger, ok := tt.params.Logger.(*logger.MockLogger); ok {
				mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
			}

			crawler, err := crawler.ProvideCrawler(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, crawler)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, crawler)
			}
		})
	}
}

func TestProvideCollyDebugger(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	debugger := logger.NewCollyDebugger(mockLogger)
	assert.NotNil(t, debugger)
}
