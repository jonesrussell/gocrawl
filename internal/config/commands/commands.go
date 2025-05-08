// Package commands provides constants and utilities for command handling in the application.
package commands

// Command names
const (
	// Root commands
	Indices = "index"
	Job     = "job"
	Crawl   = "crawl"
	HTTPD   = "httpd"
	Search  = "search"
	Sources = "sources"

	// Indices subcommands
	IndicesList   = "index list"
	IndicesDelete = "index delete"
	IndicesCreate = "index create"
)

// ConfigRequirements represents which configuration sections are required for a command
type ConfigRequirements struct {
	// RequiresCrawler indicates if crawler configuration is required
	RequiresCrawler bool
	// RequiresSources indicates if sources configuration is required
	RequiresSources bool
	// RequiresElasticsearchIndex indicates if an Elasticsearch index name is required
	RequiresElasticsearchIndex bool
}

// GetConfigRequirements returns the configuration requirements for a given command
func GetConfigRequirements(command string) ConfigRequirements {
	switch command {
	case IndicesList:
		// List command only needs basic Elasticsearch connection settings, no index name
		return ConfigRequirements{
			RequiresCrawler:            false,
			RequiresSources:            false,
			RequiresElasticsearchIndex: false,
		}
	case IndicesDelete:
		// Delete command needs Elasticsearch connection but not crawler or sources
		return ConfigRequirements{
			RequiresCrawler:            false,
			RequiresSources:            false,
			RequiresElasticsearchIndex: false,
		}
	default:
		// By default, require all configurations
		return ConfigRequirements{
			RequiresCrawler:            true,
			RequiresSources:            true,
			RequiresElasticsearchIndex: true,
		}
	}
}

// RequiresCrawlerConfig returns true if the given command requires crawler configuration.
// Some commands like index list don't need crawler config to function.
func RequiresCrawlerConfig(command string) bool {
	return GetConfigRequirements(command).RequiresCrawler
}

// RequiresSourcesConfig returns true if the given command requires sources configuration.
func RequiresSourcesConfig(command string) bool {
	return GetConfigRequirements(command).RequiresSources
}

// RequiresElasticsearchIndex returns true if the given command requires an Elasticsearch index name.
func RequiresElasticsearchIndex(command string) bool {
	return GetConfigRequirements(command).RequiresElasticsearchIndex
}
