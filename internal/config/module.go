package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// New creates a new Config instance with values from Viper
func New() (*Config, error) {
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

// Module provides the config module and its dependencies
var Module = fx.Options(
	fx.Provide(
		New,              // Provide the New function to return *Config
		NewHTTPTransport, // Ensure this is also provided if needed
	),
)
