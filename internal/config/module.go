package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// InitializeConfig sets up the configuration
func InitializeConfig(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	// Set default values
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("APP_ENV", "development")

	// Bind environment variables and check for errors
	if err := viper.BindEnv("LOG_LEVEL"); err != nil {
		return nil, fmt.Errorf("failed to bind LOG_LEVEL environment variable: %w", err)
	}
	if err := viper.BindEnv("APP_ENV"); err != nil {
		return nil, fmt.Errorf("failed to bind APP_ENV environment variable: %w", err)
	}

	return New()
}

// New creates a new Config instance with values from Viper
func New() (*Config, error) {
	// Set config defaults if not already configured
	if viper.ConfigFileUsed() == "" {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configErr *viper.ConfigFileNotFoundError
		if errors.As(err, &configErr) {
			// Log to stderr instead of using fmt.Println
			fmt.Fprintf(os.Stderr, "Config file not found; using environment variables\n")
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Proceed to read the configuration values
	rateLimit, err := parseRateLimit(viper.GetString(CrawlerRateLimitKey))
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Environment: viper.GetString(AppEnvKey),
			Name:        viper.GetString("APP_NAME"),
			Version:     viper.GetString("APP_VERSION"),
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
		return nil, fmt.Errorf("config validation failed: %w", validateErr)
	}

	return cfg, nil
}

// Module provides the config module and its dependencies
var Module = fx.Options(
	fx.Provide(
		New,              // Provide the New function to return *Config
		NewHTTPTransport, // Ensure this is also provided if needed
	),
)
