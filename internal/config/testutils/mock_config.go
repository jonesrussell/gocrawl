package testutils

import (
	"time"

	"crypto/tls"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/stretchr/testify/mock"
)

const (
	DefaultHTTPPort       = 8080
	DefaultReadTimeout    = 15 * time.Second
	DefaultWriteTimeout   = 15 * time.Second
	DefaultIdleTimeout    = 90 * time.Second
	DefaultMaxHeaderBytes = 1 << 20
	DefaultMaxDepth       = 3
	DefaultMaxConcurrency = 2
	DefaultRequestTimeout = 30 * time.Second
	DefaultDelay          = 2 * time.Second
	DefaultRandomDelay    = 500 * time.Millisecond
	DefaultPriority       = 5
	DefaultMaxPriority    = 10
	DefaultMaxRetries     = 3
	DefaultBulkSize       = 1000
	DefaultFlushInterval  = 30 * time.Second
	// DefaultMaxWait is the default maximum wait time for operations.
	DefaultMaxWait = 5 * time.Second
)

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {
	mock.Mock
	AppConfig           *app.Config
	LogConfig           *log.Config
	ServerConfig        *server.Config
	Sources             []types.Source
	CrawlerConfig       *crawler.Config
	PriorityConfig      *priority.Config
	ElasticsearchConfig *elasticsearch.Config
	StorageConfig       *storage.Config
	Command             string
	ValidateError       error
	ConfigFile          string
}

// GetSources implements config.Interface.
func (m *MockConfig) GetSources() []types.Source {
	args := m.Called()
	sources, _ := args.Get(0).([]types.Source)
	return sources
}

// GetCrawlerConfig implements config.Interface.
func (m *MockConfig) GetCrawlerConfig() *crawler.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*crawler.Config)
	return cfg
}

// GetLogConfig implements config.Interface.
func (m *MockConfig) GetLogConfig() *log.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*log.Config)
	return cfg
}

// GetAppConfig implements config.Interface.
func (m *MockConfig) GetAppConfig() *app.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*app.Config)
	return cfg
}

// GetElasticsearchConfig implements config.Interface.
func (m *MockConfig) GetElasticsearchConfig() *elasticsearch.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*elasticsearch.Config)
	return cfg
}

// GetServerConfig implements config.Interface.
func (m *MockConfig) GetServerConfig() *server.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*server.Config)
	return cfg
}

// GetCommand implements config.Interface.
func (m *MockConfig) GetCommand() string {
	args := m.Called()
	if args.Get(0) == nil {
		return "test"
	}
	return args.String(0)
}

// GetPriorityConfig implements config.Interface.
func (m *MockConfig) GetPriorityConfig() *priority.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*priority.Config)
	return cfg
}

// GetStorageConfig implements config.Interface.
func (m *MockConfig) GetStorageConfig() *storage.Config {
	args := m.Called()
	cfg, _ := args.Get(0).(*storage.Config)
	return cfg
}

// Validate implements config.Interface.
func (m *MockConfig) Validate() error {
	return m.ValidateError
}

// GetConfigFile returns the path to the config file.
func (m *MockConfig) GetConfigFile() string {
	return m.ConfigFile
}

// Ensure MockConfig implements config.Interface
var _ config.Interface = (*MockConfig)(nil)

// NewMockConfig creates a new mock configuration for testing.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		AppConfig: &app.Config{
			Name:        "gocrawl",
			Version:     "test",
			Environment: "test",
			Debug:       false,
		},
		LogConfig: &log.Config{
			Format: "json",
			Level:  "info",
			Output: "stdout",
		},
		ServerConfig: &server.Config{
			Host:           "localhost",
			Port:           DefaultHTTPPort,
			ReadTimeout:    DefaultReadTimeout,
			WriteTimeout:   DefaultWriteTimeout,
			IdleTimeout:    DefaultIdleTimeout,
			MaxHeaderBytes: DefaultMaxHeaderBytes,
			TLS: struct {
				Enabled  bool   `yaml:"enabled"`
				CertFile string `yaml:"cert_file"`
				KeyFile  string `yaml:"key_file"`
			}{
				Enabled: false,
			},
		},
		Sources: []types.Source{
			{
				Name: "test",
				URL:  "http://test.com",
			},
		},
		CrawlerConfig: &crawler.Config{
			MaxDepth:         DefaultMaxDepth,
			MaxConcurrency:   DefaultMaxConcurrency,
			RequestTimeout:   DefaultRequestTimeout,
			UserAgent:        "gocrawl/1.0",
			RespectRobotsTxt: true,
			AllowedDomains:   []string{"*"},
			Delay:            DefaultDelay,
			RandomDelay:      DefaultRandomDelay,
			SourceFile:       "sources.yml",
			Debug:            false,
			TLS: crawler.TLSConfig{
				InsecureSkipVerify:       false,
				MinVersion:               tls.VersionTLS12,
				MaxVersion:               0,
				PreferServerCipherSuites: true,
			},
		},
		PriorityConfig: &priority.Config{
			DefaultPriority:   DefaultPriority,
			MaxPriority:       DefaultMaxPriority,
			MinPriority:       1,
			PriorityIncrement: 1,
			PriorityDecrement: 1,
			Rules: []priority.Rule{
				{
					Pattern:  ".*",
					Priority: DefaultPriority,
				},
			},
		},
		ElasticsearchConfig: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
			Username:  "elastic",
			Password:  "elastic",
			IndexName: "gocrawl",
			Retry: struct {
				Enabled     bool          `yaml:"enabled"`
				InitialWait time.Duration `yaml:"initial_wait"`
				MaxWait     time.Duration `yaml:"max_wait"`
				MaxRetries  int           `yaml:"max_retries"`
			}{
				Enabled:     true,
				InitialWait: 1 * time.Second,
				MaxWait:     DefaultMaxWait,
				MaxRetries:  DefaultMaxRetries,
			},
			BulkSize:      DefaultBulkSize,
			FlushInterval: DefaultFlushInterval,
		},
		StorageConfig: &storage.Config{
			Type: "elasticsearch",
		},
		Command:       "test",
		ValidateError: nil,
		ConfigFile:    "config.yaml",
	}
}
