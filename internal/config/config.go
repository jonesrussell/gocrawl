package config

import (
	"errors"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
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

	// Log the full configuration if debug is enabled
	if cfg.App.Debug {
		logConfig(cfg)
	}

	return cfg, nil
}

// logConfig logs the configuration values
func logConfig(cfg *Config) {
	// Create a logger instance (you can customize this as needed)
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {

		}
	}(logger) // flushes buffer, if any
	logger.Debug("Loaded configuration",
		zap.String("Environment", cfg.App.Environment),
		zap.String("LogLevel", cfg.App.LogLevel),
		zap.Bool("Debug", cfg.App.Debug),
		zap.String("ElasticsearchURL", cfg.Elasticsearch.URL),
		zap.String("ElasticsearchUsername", cfg.Elasticsearch.Username),
		zap.String("ElasticsearchAPIKey", cfg.Elasticsearch.APIKey),
		zap.String("CrawlerBaseURL", cfg.Crawler.BaseURL),
		zap.Int("CrawlerMaxDepth", cfg.Crawler.MaxDepth),
		zap.Duration("CrawlerRateLimit", cfg.Crawler.RateLimit),
		zap.String("ElasticsearchPassword", cfg.Elasticsearch.Password), // Be cautious with sensitive data
	)
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}
