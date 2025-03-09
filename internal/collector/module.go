// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector module and its dependencies.
// It uses fx.Module to define the collector package as a dependency injection module,
// providing the New function as a constructor for creating collector instances.
//
// The module is used to integrate the collector package into the main application
// using the fx dependency injection framework.
func Module() fx.Option {
	return fx.Module("collector",
		fx.Provide(
			New,
		),
	)
}
