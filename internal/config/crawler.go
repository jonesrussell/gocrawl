// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/viper"
)

// createCrawlerConfig creates the crawler configuration
func createCrawlerConfig() (CrawlerConfig, error) {
	fmt.Printf("DEBUG: Creating crawler config. Environment: %s\n", viper.GetString("app.environment"))

	rateLimit, err := parseRateLimit(viper.GetString("crawler.rate_limit"))
	if err != nil {
		fmt.Printf("DEBUG: Failed to parse rate limit: %v\n", err)
		return CrawlerConfig{}, fmt.Errorf("error parsing rate limit: %w", err)
	}
	fmt.Printf("DEBUG: Rate limit parsed successfully: %v\n", rateLimit)

	sourceFile := viper.GetString("crawler.source_file")
	if sourceFile == "" {
		fmt.Printf("DEBUG: Source file is empty\n")
		return CrawlerConfig{}, fmt.Errorf("source file is required")
	}
	fmt.Printf("DEBUG: Source file: %s\n", sourceFile)

	var sources []Source
	sources, err = loadSources(sourceFile)
	if err != nil {
		fmt.Printf("DEBUG: Failed to load sources: %v\n", err)
		return CrawlerConfig{}, fmt.Errorf("failed to load sources: %w", err)
	}
	fmt.Printf("DEBUG: Sources loaded successfully: %d sources\n", len(sources))

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

// loadSources loads sources from a file
func loadSources(path string) ([]Source, error) {
	fmt.Printf("DEBUG: Loading sources from file: %s\n", path)

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
