package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestNewHTTPTransport(t *testing.T) {
	transport := config.NewHTTPTransport()
	assert.NotNil(t, transport)
}
