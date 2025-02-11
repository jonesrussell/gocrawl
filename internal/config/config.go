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
	IndexName string // Added IndexName here since it's crawler-related
	Transport http.RoundTripper
}

// ElasticsearchConfig holds Elasticsearch-specific configuration
type ElasticsearchConfig struct {
	URL      string
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
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	rateLimit := viper.GetDuration("CRAWLER_RATE_LIMIT")
	if rateLimit == 0 {
		rateLimit = time.Second
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
			IndexName: viper.GetString("INDEX_NAME"),
			Transport: transport,
		},
		Elasticsearch: ElasticsearchConfig{
			URL:      viper.GetString("ELASTIC_URL"),
			Password: viper.GetString("ELASTIC_PASSWORD"),
			APIKey:   viper.GetString("ELASTIC_API_KEY"),
		},
	}

	if cfg.Elasticsearch.URL == "" {
		return nil, ErrMissingElasticURL
	}

	return cfg, nil
}
