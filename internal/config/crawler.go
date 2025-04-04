// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/viper"
)

// createCrawlerConfig creates the crawler configuration
func createCrawlerConfig() (CrawlerConfig, error) {
	rateLimit, err := parseRateLimit(viper.GetString("crawler.rate_limit"))
	if err != nil {
		return CrawlerConfig{}, fmt.Errorf("error parsing rate limit: %w", err)
	}

	sourceFile := viper.GetString("crawler.source_file")
	if sourceFile == "" {
		return CrawlerConfig{}, fmt.Errorf("source file is required")
	}

	sources, err := loadSources(sourceFile)
	if err != nil {
		return CrawlerConfig{}, fmt.Errorf("failed to load sources: %w", err)
	}

	return CrawlerConfig{
		BaseURL:          viper.GetString("crawler.base_url"),
		MaxDepth:         viper.GetInt("crawler.max_depth"),
		RateLimit:        rateLimit,
		RandomDelay:      viper.GetDuration("crawler.random_delay"),
		IndexName:        viper.GetString("crawler.index_name"),
		ContentIndexName: viper.GetString("crawler.content_index_name"),
		SourceFile:       sourceFile,
		Parallelism:      viper.GetInt("crawler.parallelism"),
		Sources:          sources,
	}, nil
}

// loadSources loads sources from the source file
func loadSources(sourceFile string) ([]Source, error) {
	// Load sources from the source file
	sourcesConfig, err := loader.LoadFromFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources from %s: %w", sourceFile, err)
	}

	// Debug: Print the loaded config
	fmt.Printf("Loaded sources config: %+v\n", sourcesConfig)

	// Convert loader.Config to config.Source
	var sources []Source
	for i := range sourcesConfig {
		rateLimit, parseErr := ParseRateLimit(sourcesConfig[i].RateLimit)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse rate limit for source %s: %w", sourcesConfig[i].Name, parseErr)
		}

		sources = append(sources, Source{
			Name:         sourcesConfig[i].Name,
			URL:          sourcesConfig[i].URL,
			RateLimit:    rateLimit,
			MaxDepth:     sourcesConfig[i].MaxDepth,
			Time:         sourcesConfig[i].Time,
			ArticleIndex: sourcesConfig[i].ArticleIndex,
			Index:        sourcesConfig[i].Index,
			Selectors: SourceSelectors{
				Article: ArticleSelectors{
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
				},
			},
		})
	}

	return sources, nil
}
