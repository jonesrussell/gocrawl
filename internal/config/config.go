package config

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Error definitions
var (
	ErrMissingElasticURL = errors.New("ELASTIC_URL is required")
)

// Configuration keys
const (
	AppEnvKey           = "APP_ENV"
	LogLevelKey         = "LOG_LEVEL"
	AppDebugKey         = "APP_DEBUG"
	CrawlerBaseURLKey   = "CRAWLER_BASE_URL"
	CrawlerMaxDepthKey  = "CRAWLER_MAX_DEPTH"
	CrawlerRateLimitKey = "CRAWLER_RATE_LIMIT"
	ElasticURLKey       = "ELASTIC_URL"
	ElasticUsernameKey  = "ELASTIC_USERNAME"
	ElasticPasswordKey  = "ELASTIC_PASSWORD"
	ElasticAPIKeyKey    = "ELASTIC_API_KEY"
	ElasticIndexNameKey = "ELASTIC_INDEX_NAME"
	ElasticSkipTLSKey   = "ELASTIC_SKIP_TLS"
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

// Config holds all configuration settings
type Config struct {
	App           AppConfig
	Crawler       CrawlerConfig
	Elasticsearch ElasticsearchConfig
}

// NewConfig creates a new Config instance with values from Viper
func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// You can log this if needed
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
			LogLevel:    viper.GetString(LogLevelKey),
			Debug:       viper.GetBool(AppDebugKey),
		},
		Crawler: CrawlerConfig{
			BaseURL:   viper.GetString(CrawlerBaseURLKey),
			MaxDepth:  viper.GetInt(CrawlerMaxDepthKey),
			RateLimit: rateLimit,
			IndexName: viper.GetString(ElasticIndexNameKey),
		},
		Elasticsearch: ElasticsearchConfig{
			URL:       viper.GetString(ElasticURLKey),
			Username:  viper.GetString(ElasticUsernameKey),
			Password:  viper.GetString(ElasticPasswordKey),
			APIKey:    viper.GetString(ElasticAPIKeyKey),
			IndexName: viper.GetString(ElasticIndexNameKey),
			SkipTLS:   viper.GetBool(ElasticSkipTLSKey),
		},
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	// Log the full configuration if debug is enabled
	if cfg.App.Debug {
		logConfig(cfg)
	}

	return cfg, nil
}

// parseRateLimit parses the rate limit duration from a string
func parseRateLimit(rateLimitStr string) (time.Duration, error) {
	if rateLimitStr == "" {
		return time.Second, fmt.Errorf("rate limit cannot be empty")
	}
	rateLimit, err := time.ParseDuration(rateLimitStr)
	if err != nil {
		return time.Second, fmt.Errorf("error parsing duration")
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

// logConfig logs the configuration values
func logConfig(cfg *Config) {
	// Create a logger instance (you can customize this as needed)
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
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
		zap.Bool("ElasticsearchSkipTLS", cfg.Elasticsearch.SkipTLS),
	)
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// Add this function to parse rate limits
func ParseRateLimit(rate string) (time.Duration, error) {
	duration, err := time.ParseDuration(rate)
	if err != nil {
		return time.Second, nil // Return default value on error
	}
	return duration, nil
}
