package testutils

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

var setupMutex sync.Mutex

// TestSetup holds test environment configuration
type TestSetup struct {
	TmpDir      string
	ConfigPath  string
	SourcesPath string
	Cleanup     func()
}

// SetupTestEnvironment creates a test environment with config and sources files
func SetupTestEnvironment(t *testing.T, configContent string, sourcesContent string) *TestSetup {
	tmpDir := t.TempDir()
	sourcesPath := filepath.Join(tmpDir, "sources.yml")
	configPath := filepath.Join(tmpDir, "config.yml")

	setup := &TestSetup{
		TmpDir:      tmpDir,
		ConfigPath:  configPath,
		SourcesPath: sourcesPath,
	}

	// Setup sources file
	if err := setup.setupSourcesFile(sourcesContent); err != nil {
		t.Fatal(err)
	}

	// Setup config file
	if err := setup.setupConfigFile(configContent); err != nil {
		t.Fatal(err)
	}

	// Setup environment variables
	setup.setupEnvironment(t)

	return setup
}

func (s *TestSetup) setupSourcesFile(sourcesContent string) error {
	if sourcesContent == "" {
		defaultSourcesPath := "testdata/sources.yml"
		content, err := os.ReadFile(defaultSourcesPath)
		if err != nil {
			return err
		}
		sourcesContent = string(content)
	}
	return os.WriteFile(s.SourcesPath, []byte(sourcesContent), 0600)
}

func (s *TestSetup) setupConfigFile(configContent string) error {
	if configContent == "" {
		return s.setupDefaultConfig()
	}
	return s.setupCustomConfig(configContent)
}

func (s *TestSetup) setupDefaultConfig() error {
	defaultConfigPath := "testdata/config.yml"
	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return err
	}

	var config map[string]any
	if unmarshalErr := yaml.Unmarshal(content, &config); unmarshalErr != nil {
		return unmarshalErr
	}

	return s.writeConfigWithSourcePath(config)
}

func (s *TestSetup) setupCustomConfig(configContent string) error {
	var config map[string]any
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		return err
	}

	return s.writeConfigWithSourcePath(config)
}

func (s *TestSetup) writeConfigWithSourcePath(config map[string]any) error {
	// Ensure crawler section exists and has source_file
	crawlerSection, ok := config["crawler"].(map[string]any)
	if !ok {
		crawlerSection = make(map[string]any)
		config["crawler"] = crawlerSection
	}
	crawlerSection["source_file"] = s.SourcesPath
	crawlerSection["base_url"] = "http://test.example.com"

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(s.ConfigPath, configBytes, 0600)
}

func (s *TestSetup) setupEnvironment(t *testing.T) {
	envVars := map[string]string{
		"GOCRAWL_APP_ENVIRONMENT":          "test",
		"GOCRAWL_LOG_LEVEL":                "debug",
		"GOCRAWL_CRAWLER_MAX_DEPTH":        "2",
		"GOCRAWL_CRAWLER_PARALLELISM":      "2",
		"GOCRAWL_SERVER_SECURITY_ENABLED":  "true",
		"GOCRAWL_SERVER_SECURITY_API_KEY":  "id:test_api_key",
		"GOCRAWL_ELASTICSEARCH_ADDRESSES":  "http://localhost:9200",
		"GOCRAWL_ELASTICSEARCH_INDEX_NAME": "test-index",
		"GOCRAWL_ELASTICSEARCH_API_KEY":    "id:test_api_key",
		"GOCRAWL_CRAWLER_SOURCE_FILE":      s.SourcesPath,
		"GOCRAWL_CRAWLER_BASE_URL":         "http://test.example.com",
		"GOCRAWL_CONFIG_FILE":              s.ConfigPath,
	}

	for k, v := range envVars {
		t.Setenv(k, v)
	}
}
