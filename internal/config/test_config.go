// Package config provides configuration management for the application.
package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config/types"
)

const (
	// TestDataDir is the directory containing test data
	TestDataDir = "testdata"
	// ConfigsDir is the directory containing test configurations
	ConfigsDir = "configs"
)

// GetTestDataPath returns the absolute path to a file in the test data directory
func GetTestDataPath(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get caller info")
	return filepath.Join(filepath.Dir(filename), "..", "testdata", relativePath)
}

// GetTestConfigPath returns the absolute path to a config file in the test data directory
func GetTestConfigPath(t *testing.T, filename string) string {
	t.Helper()
	return GetTestDataPath(t, filename)
}

// LoadTestConfig loads a test configuration from the specified filename
func LoadTestConfig(t *testing.T, filename string) *Config {
	t.Helper()
	v := viper.New()
	v.SetConfigFile(GetTestConfigPath(t, filename))
	require.NoError(t, v.ReadInConfig(), "failed to read config file")

	// Set default values for test environment
	v.SetDefault("app.environment", "test")
	v.SetDefault("app.name", "gocrawl-test")
	v.SetDefault("app.version", "0.0.1")
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.debug", true)
	v.SetDefault("crawler.base_url", "http://test.example.com")
	v.SetDefault("crawler.max_depth", DefaultMaxDepth)
	v.SetDefault("crawler.rate_limit", "2s")
	v.SetDefault("crawler.parallelism", DefaultParallelism)

	cfg := &Config{}
	require.NoError(t, v.Unmarshal(cfg), "failed to unmarshal config")
	return cfg
}

// LoadTestSources loads test source configurations from the specified filename
func LoadTestSources(t *testing.T, filename string) []types.Source {
	t.Helper()
	v := viper.New()
	v.SetConfigFile(GetTestConfigPath(t, filename))
	require.NoError(t, v.ReadInConfig(), "failed to read config file")

	var sources []types.Source
	require.NoError(t, v.UnmarshalKey("sources", &sources), "failed to unmarshal sources")
	return sources
}
