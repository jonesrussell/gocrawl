package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Error definitions
var (
	ErrMissingElasticURL = errors.New("ELASTIC_URL is required")
)

// Configuration keys
const (
	AppEnvKey            = "APP_ENV"
	LogLevelKey          = "LOG_LEVEL"
	AppDebugKey          = "APP_DEBUG"
	CrawlerBaseURLKey    = "CRAWLER_BASE_URL"
	CrawlerMaxDepthKey   = "CRAWLER_MAX_DEPTH"
	CrawlerRateLimitKey  = "CRAWLER_RATE_LIMIT"
	CrawlerSourceFileKey = "CRAWLER_SOURCE_FILE"
	ElasticURLKey        = "ELASTIC_URL"
	ElasticUsernameKey   = "ELASTIC_USERNAME"
	ElasticPasswordKey   = "ELASTIC_PASSWORD"
	ElasticIndexNameKey  = "ELASTIC_INDEX_NAME"
	ElasticSkipTLSKey    = "ELASTIC_SKIP_TLS"
	//nolint:gosec // This is a false positive
	ElasticAPIKeyKey = "ELASTIC_API_KEY"
)

// AppConfig holds application-level configuration
type AppConfig struct {
	Environment string `yaml:"environment"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
}

// CrawlerConfig holds crawler-specific configuration
type CrawlerConfig struct {
	BaseURL          string        `yaml:"base_url"`
	MaxDepth         int           `yaml:"max_depth"`
	RateLimit        time.Duration `yaml:"rate_limit"`
	RandomDelay      time.Duration `yaml:"random_delay"`
	IndexName        string        `yaml:"index_name"`
	ContentIndexName string        `yaml:"content_index_name"`
	SourceFile       string        `yaml:"source_file"`
	Parallelism      int           `yaml:"parallelism"`
}

// SetMaxDepth sets the MaxDepth in the CrawlerConfig
func (c *CrawlerConfig) SetMaxDepth(depth int) {
	c.MaxDepth = depth
	viper.Set(CrawlerMaxDepthKey, depth) // Also set it in Viper
}

// SetRateLimit sets the RateLimit in the CrawlerConfig
func (c *CrawlerConfig) SetRateLimit(rate time.Duration) {
	c.RateLimit = rate
	viper.Set(CrawlerRateLimitKey, rate.String()) // Also set it in Viper
}

// SetBaseURL sets the BaseURL in the CrawlerConfig
func (c *CrawlerConfig) SetBaseURL(url string) {
	c.BaseURL = url
	viper.Set(CrawlerBaseURLKey, url) // Also set it in Viper
}

// SetIndexName sets the IndexName in the CrawlerConfig
func (c *CrawlerConfig) SetIndexName(index string) {
	c.IndexName = index
	viper.Set(ElasticIndexNameKey, index) // Also set it in Viper
}

// ElasticsearchConfig holds Elasticsearch-specific configuration
type ElasticsearchConfig struct {
	URL       string `yaml:"url"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	APIKey    string `yaml:"api_key"`
	IndexName string `yaml:"index_name"`
	SkipTLS   bool   `yaml:"skip_tls"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	Level string `yaml:"level"`
	Debug bool   `yaml:"debug"`
}

// Source represents a news source configuration
type Source struct {
	Name         string          `yaml:"name"`
	URL          string          `yaml:"url"`
	ArticleIndex string          `yaml:"article_index"`
	Index        string          `yaml:"index"`
	RateLimit    time.Duration   `yaml:"rate_limit"`
	MaxDepth     int             `yaml:"max_depth"`
	Time         []string        `yaml:"time"`
	Selectors    SourceSelectors `yaml:"selectors"`
}

// SourceSelectors defines the selectors for a source
type SourceSelectors struct {
	Article ArticleSelectors `yaml:"article"`
}

// Config represents the application configuration
type Config struct {
	App           AppConfig           `yaml:"app"`
	Crawler       CrawlerConfig       `yaml:"crawler"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Log           LogConfig           `yaml:"log"`
	Sources       []Source            `yaml:"sources"`
}

// parseRateLimit parses the rate limit duration from a string
func parseRateLimit(rateLimitStr string) (time.Duration, error) {
	if rateLimitStr == "" {
		return time.Second, errors.New("rate limit cannot be empty")
	}
	rateLimit, err := time.ParseDuration(rateLimitStr)
	if err != nil {
		return time.Second, errors.New("error parsing duration")
	}
	return rateLimit, nil
}

// ValidateConfig validates the configuration values
func ValidateConfig(cfg *Config) error {
	if cfg.Elasticsearch.URL == "" {
		return ErrMissingElasticURL
	}
	return nil
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// ParseRateLimit parses a rate limit string and returns a time.Duration.
// If the input is invalid, it returns a default value of 1 second and an error.
func ParseRateLimit(rateLimit string) (time.Duration, error) {
	if rateLimit == "" {
		return time.Second, nil // Return default value
	}
	duration, err := time.ParseDuration(rateLimit)
	if err != nil {
		return time.Second, errors.New("error parsing duration") // Return an error message
	}
	return duration, nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if unmarshalErr := yaml.Unmarshal(data, &config); unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", unmarshalErr)
	}

	return &config, nil
}
