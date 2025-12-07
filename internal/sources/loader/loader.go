// Package loader provides functionality for loading source configurations from files.
package loader

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var (
	// ErrNoSources indicates no sources were found in the configuration
	ErrNoSources = errors.New("no sources found in configuration")
	// ErrInvalidSourceFormat indicates the source format is invalid
	ErrInvalidSourceFormat = errors.New("invalid source format")
	// ErrMissingRequiredField indicates a required field is missing
	ErrMissingRequiredField = errors.New("missing required field")
)

// Config represents a source configuration loaded from a file.
type Config struct {
	Name         string            `mapstructure:"name"`
	URL          string            `mapstructure:"url"`
	RateLimit    any               `mapstructure:"rate_limit"` // Can be string or number
	MaxDepth     int               `mapstructure:"max_depth"`
	Time         []string          `mapstructure:"time"`
	ArticleIndex string            `mapstructure:"article_index"`
	PageIndex    string            `mapstructure:"page_index"`
	Index        string            `mapstructure:"index"`
	Selectors    SourceSelectors   `mapstructure:"selectors"`
	UserAgent    string            `mapstructure:"user_agent"`
	Headers      map[string]string `mapstructure:"headers"`
}

// SourceSelectors defines the selectors for a source.
type SourceSelectors struct {
	Article ArticleSelectors `mapstructure:"article"`
	List    ListSelectors    `mapstructure:"list"`
	Page    PageSelectors    `mapstructure:"page"`
}

// ListSelectors defines the CSS selectors used for article list page extraction.
type ListSelectors struct {
	Container       string   `mapstructure:"container"`
	ArticleCards    string   `mapstructure:"article_cards"`
	ArticleList     string   `mapstructure:"article_list"`
	ExcludeFromList []string `mapstructure:"exclude_from_list"`
}

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors struct {
	Container     string   `mapstructure:"container"`
	Title         string   `mapstructure:"title"`
	Body          string   `mapstructure:"body"`
	Intro         string   `mapstructure:"intro"`
	Link          string   `mapstructure:"link"`
	Image         string   `mapstructure:"image"`
	Byline        string   `mapstructure:"byline"`
	PublishedTime string   `mapstructure:"published_time"`
	TimeAgo       string   `mapstructure:"time_ago"`
	JSONLD        string   `mapstructure:"json_ld"`
	Section       string   `mapstructure:"section"`
	Keywords      string   `mapstructure:"keywords"`
	Description   string   `mapstructure:"description"`
	OGTitle       string   `mapstructure:"og_title"`
	OGDescription string   `mapstructure:"og_description"`
	OGImage       string   `mapstructure:"og_image"`
	OGType        string   `mapstructure:"og_type"`
	OGSiteName    string   `mapstructure:"og_site_name"`
	OgURL         string   `mapstructure:"og_url"`
	Canonical     string   `mapstructure:"canonical"`
	WordCount     string   `mapstructure:"word_count"`
	PublishDate   string   `mapstructure:"publish_date"`
	Category      string   `mapstructure:"category"`
	Tags          string   `mapstructure:"tags"`
	Author        string   `mapstructure:"author"`
	BylineName    string   `mapstructure:"byline_name"`
	ArticleID     string   `mapstructure:"article_id"`
	Exclude       []string `mapstructure:"exclude"`
}

// PageSelectors defines the CSS selectors used for page content extraction.
type PageSelectors struct {
	Container     string   `mapstructure:"container"`
	Title         string   `mapstructure:"title"`
	Content       string   `mapstructure:"content"`
	Description   string   `mapstructure:"description"`
	Keywords      string   `mapstructure:"keywords"`
	OGTitle       string   `mapstructure:"og_title"`
	OGDescription string   `mapstructure:"og_description"`
	OGImage       string   `mapstructure:"og_image"`
	OgURL         string   `mapstructure:"og_url"`
	Canonical     string   `mapstructure:"canonical"`
	Exclude       []string `mapstructure:"exclude"`
}

// Loader handles loading and validating source configurations.
type Loader struct {
	configPath string
	viper      *viper.Viper
}

// NewLoader creates a new Loader instance.
func NewLoader(configPath string) (*Loader, error) {
	v, err := newConfigLoader(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	return &Loader{
		configPath: configPath,
		viper:      v,
	}, nil
}

// LoadSources loads and validates all sources from the configuration.
func (l *Loader) LoadSources() ([]Config, error) {
	sources, err := l.loadRawSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load raw sources: %w", err)
	}

	configs, err := l.validateAndConvertSources(sources)
	if err != nil {
		return nil, fmt.Errorf("failed to validate sources: %w", err)
	}

	return configs, nil
}

// loadRawSources loads the raw source data from the configuration.
func (l *Loader) loadRawSources() ([]map[string]any, error) {
	if !l.viper.IsSet("sources") {
		return nil, ErrNoSources
	}

	sourcesRaw := l.viper.Get("sources")
	sourcesArray, ok := sourcesRaw.([]any)
	if !ok {
		return nil, ErrInvalidSourceFormat
	}

	sources := make([]map[string]any, 0, len(sourcesArray))
	for _, src := range sourcesArray {
		srcMap, srcOk := src.(map[string]any)
		if !srcOk {
			continue
		}
		sources = append(sources, srcMap)
	}

	return sources, nil
}

// validateAndConvertSources validates and converts the sources to Config structs.
func (l *Loader) validateAndConvertSources(sources []map[string]any) ([]Config, error) {
	if len(sources) == 0 {
		return nil, ErrNoSources
	}

	configs := make([]Config, 0, len(sources))
	for _, src := range sources {
		cfg, convertErr := l.convertToConfig(src)
		if convertErr != nil {
			continue
		}
		if validateErr := l.validateConfig(&cfg); validateErr != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	if len(configs) == 0 {
		return nil, ErrNoSources
	}

	return configs, nil
}

// convertToConfig converts a raw source map to a Config struct.
func (l *Loader) convertToConfig(src map[string]any) (Config, error) {
	var cfg Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		return Config{}, fmt.Errorf("failed to create decoder: %w", err)
	}

	if decodeErr := decoder.Decode(src); decodeErr != nil {
		return Config{}, fmt.Errorf("failed to decode source: %w", decodeErr)
	}

	return cfg, nil
}

// validateConfig validates a source configuration.
func (l *Loader) validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	if cfg.Name == "" {
		return fmt.Errorf("%w: name", ErrMissingRequiredField)
	}

	if cfg.URL == "" {
		return fmt.Errorf("%w: url", ErrMissingRequiredField)
	}

	if err := l.validateURL(cfg.URL); err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if err := l.validateRateLimit(cfg); err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	l.validateMaxDepth(cfg)

	if err := l.validateTime(cfg); err != nil {
		return fmt.Errorf("invalid time: %w", err)
	}

	return nil
}

// validateURL validates the URL format.
func (l *Loader) validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("must be a valid HTTP(S) URL")
	}
	return nil
}

// validateRateLimit validates and normalizes the rate limit.
func (l *Loader) validateRateLimit(cfg *Config) error {
	if cfg.RateLimit == nil {
		cfg.RateLimit = "1s" // Default rate limit
		return nil
	}

	switch v := cfg.RateLimit.(type) {
	case string:
		if v == "" {
			return errors.New("cannot be empty")
		}
		if _, err := time.ParseDuration(v); err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
	case int, int64, float64:
		cfg.RateLimit = fmt.Sprintf("%ds", v)
	default:
		return errors.New("must be a string or number")
	}

	return nil
}

// validateMaxDepth validates the max depth.
func (l *Loader) validateMaxDepth(cfg *Config) {
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 2 // Default max depth
	}
}

// validateTime validates the time format.
func (l *Loader) validateTime(cfg *Config) error {
	for _, t := range cfg.Time {
		if _, err := time.Parse("15:04", t); err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}
	}
	return nil
}

// newConfigLoader creates a new Viper instance for loading configuration.
func newConfigLoader(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// If config file is not found, create a default config
		v.Set("sources", []map[string]any{})
	}

	return v, nil
}
