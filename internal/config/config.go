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
	Environment string
}

// CrawlerConfig holds crawler-specific configuration
type CrawlerConfig struct {
	BaseURL          string
	MaxDepth         int
	RateLimit        time.Duration
	RandomDelay      time.Duration
	IndexName        string
	ContentIndexName string
	SourceFile       string
	Parallelism      int
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
	URL       string
	Username  string
	Password  string
	APIKey    string
	IndexName string
	SkipTLS   bool
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	Level string
	Debug bool
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

// Config holds all configuration settings
type Config struct {
	App           AppConfig
	Crawler       CrawlerConfig
	Elasticsearch ElasticsearchConfig
	Log           LogConfig
	Sources       []Source `yaml:"sources"`
}

// NewConfig creates a new Config instance with values from Viper
func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configErr *viper.ConfigFileNotFoundError
		if errors.As(err, &configErr) {
			//nolint:forbidigo // No logger here
			fmt.Println("Config file not found; ignoring error")
		} else {
			// Config file was found but another error was produced
			return nil, err
		}
	}

	// Proceed to read the configuration values
	rateLimit, err := parseRateLimit(viper.GetString(CrawlerRateLimitKey))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Environment: viper.GetString(AppEnvKey),
		},
		Crawler: CrawlerConfig{
			BaseURL:          viper.GetString(CrawlerBaseURLKey),
			MaxDepth:         viper.GetInt(CrawlerMaxDepthKey),
			RateLimit:        rateLimit,
			RandomDelay:      viper.GetDuration("CRAWLER_RANDOM_DELAY"),
			IndexName:        viper.GetString(ElasticIndexNameKey),
			ContentIndexName: viper.GetString("ELASTIC_CONTENT_INDEX_NAME"),
			SourceFile:       viper.GetString(CrawlerSourceFileKey),
			Parallelism:      viper.GetInt("CRAWLER_PARALLELISM"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL:       viper.GetString(ElasticURLKey),
			Username:  viper.GetString(ElasticUsernameKey),
			Password:  viper.GetString(ElasticPasswordKey),
			APIKey:    viper.GetString(ElasticAPIKeyKey),
			IndexName: viper.GetString(ElasticIndexNameKey),
			SkipTLS:   viper.GetBool(ElasticSkipTLSKey),
		},
		Log: LogConfig{
			Level: viper.GetString(LogLevelKey),
			Debug: viper.GetBool(AppDebugKey),
		},
	}

	if validateErr := ValidateConfig(cfg); validateErr != nil {
		return nil, validateErr
	}

	return cfg, nil
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
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
