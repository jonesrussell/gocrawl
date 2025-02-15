package storage_test

import (
	"net/http"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := storage.DefaultOptions()

	assert.Equal(t, "5m", opts.ScrollDuration)
	assert.Empty(t, opts.Username)
	assert.Empty(t, opts.Password)
	assert.Empty(t, opts.APIKey)
	assert.Empty(t, opts.URL)
	assert.Equal(t, http.DefaultTransport, opts.Transport)
}

func TestOptionsCustomValues(t *testing.T) {
	customTransport := &http.Transport{}
	opts := storage.Options{
		URL:            "http://localhost:9200",
		Username:       "custom_user",
		Password:       "custom_pass",
		APIKey:         "custom_key",
		ScrollDuration: "1m",
		Transport:      customTransport,
	}

	assert.Equal(t, "http://localhost:9200", opts.URL)
	assert.Equal(t, "custom_user", opts.Username)
	assert.Equal(t, "custom_pass", opts.Password)
	assert.Equal(t, "custom_key", opts.APIKey)
	assert.Equal(t, "1m", opts.ScrollDuration)
	assert.Equal(t, customTransport, opts.Transport)
}

func TestOptionsEmptyValues(t *testing.T) {
	opts := storage.Options{}

	assert.Empty(t, opts.URL)
	assert.Empty(t, opts.Username)
	assert.Empty(t, opts.Password)
	assert.Empty(t, opts.APIKey)
	assert.Empty(t, opts.ScrollDuration)
	assert.Nil(t, opts.Transport)
}
