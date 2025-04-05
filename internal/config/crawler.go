// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/viper"
)

// createCrawlerConfig creates the crawler configuration
func createCrawlerConfig(v *viper.Viper) (CrawlerConfig, error) {
	fmt.Printf("DEBUG: Creating crawler config. Environment: %s\n", v.GetString("app.environment"))

	rateLimit, err := parseRateLimit(v.GetString("crawler.rate_limit"))
	if err != nil {
		fmt.Printf("DEBUG: Failed to parse rate limit: %v\n", err)
		return CrawlerConfig{}, fmt.Errorf("error parsing rate limit: %w", err)
	}
	fmt.Printf("DEBUG: Rate limit parsed successfully: %v\n", rateLimit)

	sourceFile := v.GetString("crawler.source_file")
	if sourceFile == "" {
		fmt.Printf("DEBUG: Source file is empty\n")
		return CrawlerConfig{}, fmt.Errorf("source file is required")
	}
	fmt.Printf("DEBUG: Source file: %s\n", sourceFile)

	return CrawlerConfig{
		BaseURL:          v.GetString("crawler.base_url"),
		MaxDepth:         v.GetInt("crawler.max_depth"),
		RateLimit:        rateLimit,
		RandomDelay:      v.GetDuration("crawler.random_delay"),
		IndexName:        v.GetString("crawler.index_name"),
		ContentIndexName: v.GetString("crawler.content_index_name"),
		SourceFile:       sourceFile,
		Parallelism:      v.GetInt("crawler.parallelism"),
	}, nil
}

// loadSources loads sources from a file
func loadSources(path string) ([]Source, error) {
	fmt.Printf("DEBUG: Loading sources from file: %s\n", path)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %s", path)
	}

	sourcesConfig, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	fmt.Printf("DEBUG: Loaded sources config: %+v\n", sourcesConfig)

	var sources []Source
	for i := range sourcesConfig {
		rateLimit, err := parseRateLimit(sourcesConfig[i].RateLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rate limit for source %s: %w", sourcesConfig[i].Name, err)
		}

		sources = append(sources, Source{
			Name:         sourcesConfig[i].Name,
			URL:          sourcesConfig[i].URL,
			RateLimit:    rateLimit,
			MaxDepth:     sourcesConfig[i].MaxDepth,
			Time:         sourcesConfig[i].Time,
			ArticleIndex: sourcesConfig[i].ArticleIndex,
			Index:        sourcesConfig[i].Index,
			Selectors: SourceSelectors{Article: ArticleSelectors{
				Container:     sourcesConfig[i].Selectors.Article.Container,
				Title:         sourcesConfig[i].Selectors.Article.Title,
				Body:          sourcesConfig[i].Selectors.Article.Body,
				Intro:         sourcesConfig[i].Selectors.Article.Intro,
				Byline:        sourcesConfig[i].Selectors.Article.Byline,
				PublishedTime: sourcesConfig[i].Selectors.Article.PublishedTime,
				TimeAgo:       sourcesConfig[i].Selectors.Article.TimeAgo,
				JSONLD:        sourcesConfig[i].Selectors.Article.JSONLD,
				Section:       sourcesConfig[i].Selectors.Article.Section,
				Keywords:      sourcesConfig[i].Selectors.Article.Keywords,
				Description:   sourcesConfig[i].Selectors.Article.Description,
				OGTitle:       sourcesConfig[i].Selectors.Article.OGTitle,
				OGDescription: sourcesConfig[i].Selectors.Article.OGDescription,
				OGImage:       sourcesConfig[i].Selectors.Article.OGImage,
				OgURL:         sourcesConfig[i].Selectors.Article.OgURL,
				Canonical:     sourcesConfig[i].Selectors.Article.Canonical,
				WordCount:     sourcesConfig[i].Selectors.Article.WordCount,
				PublishDate:   sourcesConfig[i].Selectors.Article.PublishDate,
				Category:      sourcesConfig[i].Selectors.Article.Category,
				Tags:          sourcesConfig[i].Selectors.Article.Tags,
				Author:        sourcesConfig[i].Selectors.Article.Author,
				BylineName:    sourcesConfig[i].Selectors.Article.BylineName,
			}},
		})
	}

	return sources, nil
}
