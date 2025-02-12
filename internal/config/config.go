package config

import (
	"errors"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// Error definitions
var (
	ErrMissingElasticURL = errors.New("ELASTIC_URL is required")
)

// AppConfig holds application-level configuration
type AppConfig struct {
	Environment string
	LogLevel    string
	Debug       bool
}

// CrawlerConfig holds crawler-specific configuration
type CrawlerConfig struct {
	BaseURL   string
	MaxDepth  int
	RateLimit time.Duration
	IndexName string
	Transport http.RoundTripper
	SkipTLS   bool
}

// ElasticsearchConfig holds Elasticsearch-specific configuration
type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
	APIKey   string
}

// Config holds all configuration settings
type Config struct {
	App           AppConfig
	Crawler       CrawlerConfig
	Elasticsearch ElasticsearchConfig
}

// NewConfig creates a new Config instance with values from Viper
func NewConfig(transport http.RoundTripper) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	rateLimitStr := viper.GetString("CRAWLER_RATE_LIMIT")
	rateLimit, err := time.ParseDuration(rateLimitStr)
	if err != nil {
		rateLimit = time.Second // Default value if parsing fails
	}

	cfg := &Config{
		App: AppConfig{
			Environment: viper.GetString("APP_ENV"),
			LogLevel:    viper.GetString("LOG_LEVEL"),
			Debug:       viper.GetBool("APP_DEBUG"),
		},
		Crawler: CrawlerConfig{
			BaseURL:   viper.GetString("CRAWLER_BASE_URL"),
			MaxDepth:  viper.GetInt("CRAWLER_MAX_DEPTH"),
			RateLimit: rateLimit,
			IndexName: viper.GetString("ELASTIC_INDEX_NAME"),
			Transport: transport,
			SkipTLS:   viper.GetBool("CRAWLER_ELASTIC_SKIP_TLS"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL:      viper.GetString("ELASTIC_URL"),
			Username: viper.GetString("ELASTIC_USERNAME"),
			Password: viper.GetString("ELASTIC_PASSWORD"),
			APIKey:   viper.GetString("ELASTIC_API_KEY"),
		},
	}

	if cfg.Elasticsearch.URL == "" {
		return nil, ErrMissingElasticURL
	}

	return cfg, nil
}

// LoadConfig loads the configuration from environment variables or a config file
// Keeping this function if needed for other purposes
func LoadConfig() (*Config, error) {
	return &Config{
		Crawler: CrawlerConfig{
			IndexName: viper.GetString("CRAWLER_ELASTIC_INDEX_NAME"),
			BaseURL:   viper.GetString("CRAWLER_BASE_URL"),
			MaxDepth:  viper.GetInt("CRAWLER_MAX_DEPTH"),
			RateLimit: viper.GetDuration("CRAWLER_RATE_LIMIT"),
			SkipTLS:   viper.GetBool("ELASTIC_SKIP_TLS"),
		},
	}, nil
}
